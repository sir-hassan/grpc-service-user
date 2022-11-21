package main

import (
	"context"
	"fmt"
	"strconv"
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

	// add 10 users
	var ids []string
	for i := 0; i < 10; i++ {
		u, err := client.AddUser(context.Background(), &api.AddUserRequest{
			FirstName: "user_first_name_" + strconv.Itoa(i),
			LastName:  "user_last_name_" + strconv.Itoa(i),
			Country:   "country" + strconv.Itoa(i),
		})
		if err != nil {
			lg.Fatal().Err(err).Msg("call add user")
		}
		ids = append(ids, u.Id)
	}
	lg.Info().Msg("✅ adding 10 users")

	// update users
	for i := 0; i < 10; i++ {
		updatedName := "user_updated_first_name"
		_, err = client.UpdateUser(context.Background(), &api.UpdateUserRequest{
			Id:        ids[i],
			FirstName: &updatedName,
		})
		if err != nil {
			lg.Fatal().Err(err).Msg("call update user")
		}
	}
	lg.Info().Msg("✅ updating 10 users")

	// delete users
	for i := 0; i < 10; i++ {
		_, err = client.DeleteUser(context.Background(), &api.DeleteUserRequest{
			Id: ids[i],
		})
		if err != nil {
			lg.Fatal().Err(err).Msg("call delete user")
		}
	}
	lg.Info().Msg("✅ deleting 10 users")
}
