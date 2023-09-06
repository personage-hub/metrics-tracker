package logger

import (
	"go.uber.org/zap"
)

func Initialize(level string) (*zap.Logger, error) {

	var log = zap.NewNop()

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	log = zl
	return log, nil
}
