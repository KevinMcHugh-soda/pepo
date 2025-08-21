package logging

import "go.uber.org/zap"

// Init sets up a global zap logger and returns it.
func Init() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(logger)
	return logger, nil
}
