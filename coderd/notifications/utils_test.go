package notifications_test

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/slogtest"
	"github.com/coder/serpent"

	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/dbauthz"
	"github.com/coder/coder/v2/coderd/database/dbtestutil"
	"github.com/coder/coder/v2/coderd/database/pubsub"
	"github.com/coder/coder/v2/coderd/notifications"
	"github.com/coder/coder/v2/coderd/notifications/dispatch"
	"github.com/coder/coder/v2/coderd/notifications/types"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/testutil"
)

func setup(t *testing.T) (context.Context, slog.Logger, database.Store, *pubsub.PGPubsub) {
	t.Helper()

	connectionURL, closeFunc, err := dbtestutil.Open()
	require.NoError(t, err)
	t.Cleanup(closeFunc)

	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitSuperLong)
	t.Cleanup(cancel)
	logger := slogtest.Make(t, &slogtest.Options{IgnoreErrors: true, IgnoredErrorIs: []error{}}).Leveled(slog.LevelDebug)

	sqlDB, err := sql.Open("postgres", connectionURL)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	db := database.New(sqlDB)
	ps, err := pubsub.New(ctx, logger, sqlDB, connectionURL)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, ps.Close())
	})

	// nolint:gocritic // unit tests.
	return dbauthz.AsSystemRestricted(ctx), logger, db, ps
}

func defaultNotificationsConfig(method database.NotificationMethod) codersdk.NotificationsConfig {
	return codersdk.NotificationsConfig{
		Method:              serpent.String(method),
		MaxSendAttempts:     5,
		RetryInterval:       serpent.Duration(time.Millisecond * 50),
		StoreSyncInterval:   serpent.Duration(time.Millisecond * 200),
		StoreSyncBufferSize: 50,
		LeasePeriod:         serpent.Duration(time.Second * 10),
		LeaseCount:          10,
		FetchInterval:       serpent.Duration(time.Millisecond * 100),
		DispatchTimeout:     serpent.Duration(time.Minute),
		SMTP:                codersdk.NotificationsEmailConfig{},
		Webhook:             codersdk.NotificationsWebhookConfig{},
	}
}

func defaultHelpers() map[string]any {
	return map[string]any{
		"base_url": func() string { return "http://test.com" },
	}
}

func createMetrics() *notifications.Metrics {
	return notifications.NewMetrics(prometheus.NewRegistry())
}

type dispatchInterceptor struct {
	handler notifications.Handler

	sent        atomic.Int32
	retryable   atomic.Int32
	unretryable atomic.Int32
	err         atomic.Int32
	lastErr     atomic.Value
}

func newDispatchInterceptor(h notifications.Handler) *dispatchInterceptor {
	return &dispatchInterceptor{handler: h}
}

func (i *dispatchInterceptor) Dispatcher(payload types.MessagePayload, title, body string) (dispatch.DeliveryFunc, error) {
	return func(ctx context.Context, msgID uuid.UUID) (retryable bool, err error) {
		deliveryFn, err := i.handler.Dispatcher(payload, title, body)
		if err != nil {
			return false, err
		}

		retryable, err = deliveryFn(ctx, msgID)

		if err != nil {
			i.err.Add(1)
			i.lastErr.Store(err)
		}

		switch {
		case !retryable && err == nil:
			i.sent.Add(1)
		case retryable:
			i.retryable.Add(1)
		case !retryable && err != nil:
			i.unretryable.Add(1)
		}
		return retryable, err
	}, nil
}
