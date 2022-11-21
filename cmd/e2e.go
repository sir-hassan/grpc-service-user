package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/sir-hassan/grpc-service-user/api"
	"google.golang.org/grpc"
)

func runE2eCommand(lg zerolog.Logger) {
	//nolint
	opts := []grpc.DialOption{grpc.WithInsecure()}
	serverAddr := "api:8080"
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		lg.Fatal().Err(err).Msg("dialing server failed")

		return
	}
	state := conn.GetState()
	fmt.Printf("%s\n", state.String())

	defer conn.Close()
	client := api.NewUserStoreClient(conn)

	// wait for the service to be healthy
	var healthCheck *api.CheckHealthReply
	for i := 0; i < 10; i++ {
		healthCheck, err = client.CheckHealth(context.Background(), &api.CheckHealthRequest{})
		if err == nil && healthCheck.IsHealthy {
			break
		}
		if err != nil {
			lg.Err(err).Msg("failed calling health check")
		}
		lg.Info().Msg("service not ready!")
		time.Sleep(time.Second)
	}

	if !healthCheck.IsHealthy {
		lg.Fatal().Err(err).Msg("negative health check")
	}

	lg.Info().Msg("✅ all e2e tests passed")
}
