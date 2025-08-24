package services

import (
	"io"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
)

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

func stringToPointer(s string) *string {
	return &s
}

func boolToPointer(b bool) *bool {
	return &b
}

func float64ToPointer(f float64) *float64 {
	return &f
}

// setupMockController creates a new gomock controller with cleanup
func setupMockController(t *testing.T) *gomock.Controller {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	
	return ctrl
}

