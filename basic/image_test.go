package basic

import (
	"testing"
	"time"
)

func TestZapInfo(t *testing.T) {
	defer logger.Sync()
	logger.Infow("hello",
		"time", time.Now(),
	)
}
