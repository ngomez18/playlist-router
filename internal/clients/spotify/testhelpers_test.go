package spotifyclient

import (
	"io"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
)

func setupMockController(t *testing.T) *gomock.Controller {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	return ctrl
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}
