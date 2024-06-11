package notifications_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/slogtest"
	"github.com/coder/coder/v2/coderd/coderdtest"
	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/dbmem"
	"github.com/coder/coder/v2/coderd/notifications"
	"github.com/coder/coder/v2/coderd/notifications/dispatch"
	"github.com/coder/coder/v2/coderd/notifications/types"
	"github.com/coder/coder/v2/testutil"
)

// TestSingletonRegistration tests that a Manager which has been instantiated but not registered will error.
func TestSingletonRegistration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slogtest.Make(t, &slogtest.Options{IgnoreErrors: true, IgnoredErrorIs: []error{}}).Leveled(slog.LevelDebug)

	mgr, err := notifications.NewManager(defaultNotificationsConfig(), dbmem.New(), logger, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, mgr.Stop(ctx))
	})

	// Not registered yet.
	_, err = notifications.Enqueue(ctx, uuid.New(), notifications.TemplateWorkspaceDeleted, nil, "")
	require.ErrorIs(t, err, notifications.SingletonNotRegisteredErr)

	// Works after registering.
	notifications.RegisterInstance(mgr)
	_, err = notifications.Enqueue(ctx, uuid.New(), notifications.TemplateWorkspaceDeleted, nil, "")
	require.NoError(t, err)
}

func TestBufferedUpdates(t *testing.T) {
	t.Parallel()

	// setup
	ctx, logger, db, ps := setup(t)
	interceptor := &bulkUpdateInterceptor{Store: db}

	santa := &santaHandler{}
	handlers, err := notifications.NewHandlerRegistry(santa)
	require.NoError(t, err)
	mgr, err := notifications.NewManager(defaultNotificationsConfig(), interceptor, logger.Named("notifications"), nil)
	require.NoError(t, err)
	mgr.WithHandlers(handlers)

	client := coderdtest.New(t, &coderdtest.Options{Database: db, Pubsub: ps})
	user := coderdtest.CreateFirstUser(t, client)

	// given
	if _, err := mgr.Enqueue(ctx, user.UserID, notifications.TemplateWorkspaceDeleted, types.Labels{"nice": "true"}, ""); true {
		require.NoError(t, err)
	}
	if _, err := mgr.Enqueue(ctx, user.UserID, notifications.TemplateWorkspaceDeleted, types.Labels{"nice": "false"}, ""); true {
		require.NoError(t, err)
	}

	// when
	mgr.Run(ctx, 1)

	// then

	// Wait for messages to be dispatched.
	require.Eventually(t, func() bool { return santa.naughty.Load() == 1 && santa.nice.Load() == 1 }, testutil.WaitMedium, testutil.IntervalFast)

	// Stop the manager which forces an update of buffered updates.
	require.NoError(t, mgr.Stop(ctx))

	// Wait until both success & failure updates have been sent to the store.
	require.Eventually(t, func() bool { return interceptor.failed.Load() == 1 && interceptor.sent.Load() == 1 }, testutil.WaitMedium, testutil.IntervalFast)
}

type bulkUpdateInterceptor struct {
	notifications.Store

	sent   atomic.Int32
	failed atomic.Int32
}

func (b *bulkUpdateInterceptor) BulkMarkNotificationMessagesSent(ctx context.Context, arg database.BulkMarkNotificationMessagesSentParams) (int64, error) {
	b.sent.Add(int32(len(arg.IDs)))
	return b.Store.BulkMarkNotificationMessagesSent(ctx, arg)
}

func (b *bulkUpdateInterceptor) BulkMarkNotificationMessagesFailed(ctx context.Context, arg database.BulkMarkNotificationMessagesFailedParams) (int64, error) {
	b.failed.Add(int32(len(arg.IDs)))
	return b.Store.BulkMarkNotificationMessagesFailed(ctx, arg)
}

// santaHandler only dispatches nice messages.
type santaHandler struct {
	naughty atomic.Int32
	nice    atomic.Int32
}

func (*santaHandler) NotificationMethod() database.NotificationMethod {
	return database.NotificationMethodSmtp
}

func (s *santaHandler) Dispatcher(payload types.MessagePayload, _, _ string) (dispatch.DeliveryFunc, error) {
	return func(ctx context.Context, msgID uuid.UUID) (retryable bool, err error) {
		if payload.Labels.Get("nice") != "true" {
			s.naughty.Add(1)
			return false, xerrors.New("be nice")
		}

		s.nice.Add(1)
		return false, nil
	}, nil
}