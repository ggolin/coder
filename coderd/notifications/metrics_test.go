package notifications_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/coder/serpent"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	promtest "github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/coder/coder/v2/coderd/coderdtest"
	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/dbtestutil"
	"github.com/coder/coder/v2/coderd/notifications"
	"github.com/coder/coder/v2/coderd/notifications/dispatch"
	"github.com/coder/coder/v2/coderd/notifications/types"
	"github.com/coder/coder/v2/testutil"
)

func TestMetrics(t *testing.T) {
	// setup
	if !dbtestutil.WillUsePostgres() {
		t.Skip("This test requires postgres")
	}

	ctx, logger, store, ps := setup(t)

	reg := prometheus.NewRegistry()
	metrics := notifications.NewMetrics(reg)
	client := coderdtest.New(t, &coderdtest.Options{Database: store, Pubsub: ps})
	template := notifications.TemplateWorkspaceDeleted

	const (
		method      = database.NotificationMethodSmtp
		maxAttempts = 3
	)

	// given
	cfg := defaultNotificationsConfig(method)
	cfg.MaxSendAttempts = maxAttempts
	cfg.FetchInterval = serpent.Duration(time.Millisecond * 50)
	cfg.RetryInterval = serpent.Duration(time.Millisecond * 50)
	cfg.StoreSyncInterval = serpent.Duration(time.Millisecond * 100) // Twice as long as fetch interval to ensure we catch pending updates.
	mgr, err := notifications.NewManager(cfg, store, metrics, logger.Named("manager"))
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, mgr.Stop(ctx))
	})

	handler := &fakeHandler{}
	mgr.WithHandlers(map[database.NotificationMethod]notifications.Handler{
		method: handler,
	})
	mgr.Run(ctx)

	enq, err := notifications.NewStoreEnqueuer(cfg, store, defaultHelpers(), logger.Named("enqueuer"))
	require.NoError(t, err)

	// when
	first := coderdtest.CreateFirstUser(t, client)
	_, err = enq.Enqueue(ctx, first.UserID, template, map[string]string{"type": "success"}, "test") // this will succeed
	require.NoError(t, err)
	_, err = enq.Enqueue(ctx, first.UserID, template, map[string]string{"type": "failure"}, "test2") // this will fail and retry (maxAttempts - 1) times
	require.NoError(t, err)

	var seenPendingUpdates bool

	// then
	require.Eventually(t, func() bool {
		handler.mu.RLock()
		defer handler.mu.RUnlock()

		succeeded := len(handler.succeeded)
		failed := len(handler.failed)
		dispatched := promtest.ToFloat64(metrics.DispatchedCount.WithLabelValues(string(method), template.String()))
		tempFails := promtest.ToFloat64(metrics.TempFailureCount.WithLabelValues(string(method), template.String()))
		permFails := promtest.ToFloat64(metrics.PermFailureCount.WithLabelValues(string(method), template.String()))
		retries := promtest.ToFloat64(metrics.RetryCount.WithLabelValues(string(method), template.String()))
		pendingUpdates := promtest.ToFloat64(metrics.RetryCount.WithLabelValues(string(method), template.String()))
		queuedSecs, err := histogramSum(t, metrics.QueuedSeconds, model.LabelPair{Name: "method", Value: model.LabelValue(method)})
		require.NoError(t, err)

		// pendingUpdates is a gauge, and we just want to record that we saw it as non-zero at least once.
		if !seenPendingUpdates && pendingUpdates > 0 {
			seenPendingUpdates = true
		}

		// Debug:
		// t.Logf("S:%v, F:%v, D:%v, T:%v, P:%v, R:%v, QS:%.2f", succeeded, failed, dispatched, tempFails, permFails, retries, queuedSecs)

		// NOTE: not testing PendingUpdates because it's a gauge
		return succeeded == 1 &&
			// A total of maxAttempts will be made
			failed == maxAttempts &&
			// Only 1 message will be dispatched successfully
			dispatched == 1 &&
			// 2 temp failures, on the 3rd it'll be marked permanent failure
			tempFails == maxAttempts-1 &&
			// 1 original attempts + 2 retries = maxAttempts
			retries == maxAttempts-1 &&
			// 1 permanent failure after retries exhausted
			permFails == 1 &&
			// The notifications will have been queued for a non-zero amount of time
			queuedSecs > 0 &&
			// Pending updates were recorded
			seenPendingUpdates
	}, testutil.WaitShort, testutil.IntervalFast)
}

// histogramSum retrieves the cumulative sum of all observed samples from a given histogram.
func histogramSum(t *testing.T, h *prometheus.HistogramVec, series model.LabelPair) (float64, error) {
	t.Helper()

	out := make(chan prometheus.Metric, 1)
	h.Collect(out)

	metric := <-out
	close(out)

	require.NotNil(t, metric)

	sample := &dto.Metric{}
	require.NoError(t, metric.Write(sample))

	for _, lp := range sample.Label {
		// Ensure we have the series we're explicitly looking for.
		if lp.GetName() != string(series.Name) && lp.GetValue() != string(series.Value) {
			continue
		}

		return sample.Histogram.GetSampleSum(), nil
	}

	return -1, xerrors.New("could not find given series")
}

type fakeHandler struct {
	mu                sync.RWMutex
	succeeded, failed []string
}

func (f *fakeHandler) Dispatcher(payload types.MessagePayload, _, _ string) (dispatch.DeliveryFunc, error) {
	return func(ctx context.Context, msgID uuid.UUID) (retryable bool, err error) {
		f.mu.Lock()
		defer f.mu.Unlock()

		if payload.Labels["type"] == "success" {
			f.succeeded = append(f.succeeded, msgID.String())
			return false, nil
		}

		f.failed = append(f.failed, msgID.String())
		return true, xerrors.New("oops")
	}, nil
}
