package app

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/sir-hassan/grpc-service-user/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	ID        string
	FirstName string
	LastName  string
	Country   string
}

type UserStore struct {
	api.UserStoreServer
	db       *gorm.DB
	lg       zerolog.Logger
	notifier Notifier
}

var _ api.UserStoreServer = &UserStore{}

func (s *UserStore) UpdateUser(ctx context.Context, req *api.UpdateUserRequest) (*api.UpdateUserReply, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *UserStore) DeleteUser(ctx context.Context, req *api.DeleteUserRequest) (*api.DeleteUserReply, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func NewUserStore(db *gorm.DB, notifier Notifier, lg zerolog.Logger) *UserStore {
	return &UserStore{
		db:       db,
		lg:       lg,
		notifier: notifier,
	}
}

func (s *UserStore) CheckHealth(ctx context.Context, req *api.CheckHealthRequest) (*api.CheckHealthReply, error) {
	notHealthy := &api.CheckHealthReply{IsHealthy: true}

	sqlDB, err := s.db.DB()
	if err != nil {
		return notHealthy, nil
	}

	err = sqlDB.Ping()
	if err != nil {
		return notHealthy, nil
	}

	return &api.CheckHealthReply{
		IsHealthy: true,
	}, nil
}

func (s *UserStore) AddUser(ctx context.Context, req *api.AddUserRequest) (*api.AddUserReply, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (s *UserStore) ListUsers(req *api.ListUsersRequest, lus api.UserStore_ListUsersServer) error {
	return status.Error(codes.Unimplemented, "")
}
