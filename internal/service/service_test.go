package service

import (
	"io"
	"os"
	"testing"

	"refina-transaction/config/log"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	// Initialize logger to prevent nil pointer dereference during tests.
	// Services call log.Error/log.Warn/log.Info which dereference log.Log.
	log.Log = logrus.New()
	log.Log.SetOutput(io.Discard) // suppress all log output during testing
	log.Log.SetLevel(logrus.PanicLevel)

	os.Exit(m.Run())
}
