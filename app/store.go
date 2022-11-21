package app

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/sir-hassan/grpc-service-user/api"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	ID        string
	FirstName string
	LastName  string
	Country   string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserStore struct {
	api.UserStoreServer
	db *gorm.DB
	lg zerolog.Logger
}

var _ api.UserStoreServer = &UserStore{}

func NewUserStore(db *gorm.DB, lg zerolog.Logger) *UserStore {
	return &UserStore{
		db: db,
		lg: lg,
	}
}

func (s *UserStore) CheckHealth(ctx context.Context, req *api.CheckHealthRequest) (*api.CheckHealthReply, error) {
	notHealthy := &api.CheckHealthReply{IsHealthy: true}

	sqlDB, err := s.db.DB()
	if err != nil {
		//nolint
		return notHealthy, nil
	}

	err = sqlDB.Ping()
	if err != nil {
		//nolint
		return notHealthy, nil
	}

	return &api.CheckHealthReply{
		IsHealthy: true,
	}, nil
}
