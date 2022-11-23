package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"
	"github.com/sir-hassan/grpc-service-user/api"
	"github.com/sir-hassan/grpc-service-user/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	notifierSize = 1000
)

type envVars struct {
	Port int `env:"PORT" envDefault:"8080"`

	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	PostgresPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"admin"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"password"`
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"userdb"`

	NotifierWebHooks []string `env:"NOTIFIER_WEBHOOKS" envSeparator:","`
}

func runServerCommand(lg zerolog.Logger) {
	cfg := envVars{}
	if err := env.Parse(&cfg); err != nil {
		lg.Fatal().Err(err).Msg("couldn't parse env variables")
	}
	lg.Debug().Str("env_vars", fmt.Sprintf("%+v", cfg)).Msg("calculated env vars")

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Europe/Berlin",
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB,
	)
	lg.Debug().Str("dsn", dsn).Msg("calculated postgres dns string")

	db, err := createDatabase(dsn, lg)
	if err != nil {
		lg.Fatal().Err(err).Msg("connecting to database failed")
	}
	lg.Info().Msg("connected to database")

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		lg.Fatal().Err(err).Msg("tcp listen")
	}

	lg.Info().Int("port", cfg.Port).Msg("start tcp listener")

	notifier := app.NewHTTPNotifier(lg, http.DefaultClient, cfg.NotifierWebHooks, notifierSize)

	cancelNotifierChan := make(chan any)
	doneNotifierChan := notifier.Start(cancelNotifierChan)

	store := app.NewUserStore(db, notifier, lg)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)
	api.RegisterUserStoreServer(grpcServer, store)

	// Handle process termination.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		lg.Info().Str("sig", sig.String()).Msg("signal received")
		lg.Info().Msg("terminating server...")
		grpcServer.GracefulStop()
		close(cancelNotifierChan)
	}()

	lg.Info().Msg("starting server")

	if err = grpcServer.Serve(lis); err != nil {
		lg.Fatal().Err(err).Msg("serve grpc")
	}

	lg.Info().Msg("waiting to terminate notifier")
	<-doneNotifierChan

	lg.Info().Msg("server terminated successfully")
}

func createDatabase(dsn string, lg zerolog.Logger) (*gorm.DB, error) {
	var err error
	var db *gorm.DB

	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		// db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

		if err != nil {
			lg.Info().Msg("db is not ready!")

			continue
		}

		break
	}
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&app.User{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pool.
	//nolint
	sqlDB.SetMaxIdleConns(10)
	//nolint
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
