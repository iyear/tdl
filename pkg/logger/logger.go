package logger

import (
	"go.uber.org/zap"
	"sync"
)

var (
	Logger = zap.NewNop()
	mu     sync.Mutex
)

func SetDebug(debug bool) {
	mu.Lock()
	defer mu.Unlock()
	if debug {
		Logger, _ = zap.NewDevelopment()
		return
	}
	Logger = zap.NewNop()
}
