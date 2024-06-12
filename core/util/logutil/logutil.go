package logutil

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(level zapcore.LevelEnabler, path string) *zap.Logger {
	rotate := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
		LocalTime:  true,
		Compress:   true,
	}

	writer := zapcore.AddSync(rotate)

	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	config.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(config), writer, level)
	return zap.New(core, zap.AddCaller())
}
