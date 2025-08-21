package main

import (
	"pepo/internal/config"
	"pepo/internal/database"
	"pepo/internal/handlers"
	"pepo/internal/logging"
	"pepo/internal/server"
	"pepo/internal/version"

	"go.uber.org/zap"
)

func main() {
	logger, err := logging.Init()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	versionInfo := version.Get()
	zap.L().Info("version information", zap.String("version", versionInfo.String()))

	cfg := config.Load()
	zap.L().Info("starting server", zap.String("environment", cfg.Environment), zap.String("port", cfg.Port))

	zap.L().Info("connecting to database")
	db, queries, err := database.Initialize(cfg.DatabaseURL, database.DefaultConnectionConfig())
	if err != nil {
		zap.L().Fatal("failed to initialize database", zap.Error(err))
	}
	defer func() {
		if err := database.Close(db); err != nil {
			zap.L().Error("error closing database", zap.Error(err))
		}
	}()

	zap.L().Info("initializing application handlers")
	personHandler := handlers.NewPersonHandler(queries)
	actionHandler := handlers.NewActionHandler(queries)
	combinedAPIHandler := handlers.NewCombinedAPIHandler(personHandler, actionHandler)

	zap.L().Info("setting up HTTP server")
	srv, err := server.New(cfg, combinedAPIHandler, personHandler, actionHandler)
	if err != nil {
		zap.L().Fatal("failed to create server", zap.Error(err))
	}

	zap.L().Info("server initialization complete, starting")
	if err := srv.StartWithGracefulShutdown(); err != nil {
		zap.L().Fatal("server error", zap.Error(err))
	}
}
