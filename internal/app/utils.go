package app

import (
	"runtime/debug"

	"github.com/alkurbatov/metrics-collector/internal/logging"
)

// Log if panic occurres but try to avoid program termination.
func tryRecover() {
	if p := recover(); p != nil {
		l := logging.Log.WithField("event", "panic")
		l.Error(p)
		l.Debug(string(debug.Stack()))
	}
}
