package config

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestLoadTOML(t *testing.T) {
	ctx, doCancelFunc := context.WithCancel(context.Background())
	defer func() {
		doCancelFunc()
		time.Sleep(time.Second)
	}()
	Init(ctx)
	fmt.Println(BaseConfig)
}
