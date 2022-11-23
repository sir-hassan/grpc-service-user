package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/sir-hassan/grpc-service-user/api"
	"github.com/sir-hassan/grpc-service-user/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const defaultPort = 8080

func runServerCommand(lg zerolog.Logger) {
	dsn := "host=postgres user=myadmin password=mypassword dbname=userdb port=5432 sslmode=disable TimeZone=Europe/Berlin"
	db, err := createDatabase(dsn, lg)
	if err != nil {
		lg.Fatal().Err(err).Msg("connecting to database failed")
	}
	lg.Info().Msg("connected to database")

	port := defaultPort

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		lg.Fatal().Err(err).Msg("tcp listen")
	}

	lg.Info().Int("port", port).Msg("start tcp listener")

	notifier := app.NewHTTPNotifier(lg, http.DefaultClient, []string{"https://webhook.site/49525789-a6ae-4f41-8518-4c2d3ae8f4c3"}, 100)

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
