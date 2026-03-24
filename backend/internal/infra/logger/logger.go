package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.SugaredLogger

func Init(mode string) {
	var cfg zap.Config
	if mode == "debug" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
	L = logger.Sugar()
}

func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}
