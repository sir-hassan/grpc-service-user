package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/sir-hassan/grpc-service-user/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type User struct {
	ID        string
	FirstName string
	LastName  string

	Nickname string
	Password string
	Email    string
	Country  string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserStore struct {
	api.UserStoreServer
	db       *gorm.DB
	lg       zerolog.Logger
	notifier Notifier
}

var _ api.UserStoreServer = &UserStore{}

func (s *UserStore) UpdateUser(ctx context.Context, req *api.UpdateUserRequest) (*api.UpdateUserReply, error) {
	patches := map[string]any{}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing or empty 'id' field")
	}

	if req.FirstName != nil {
		patches["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		patches["last_name"] = *req.LastName
	}
	if req.Nickname != nil {
		patches["nickname"] = *req.Nickname
	}
	if req.Password != nil {
		patches["password"] = *req.Password
	}
	if req.Email != nil {
		patches["email"] = *req.Email
	}
	if req.Country != nil {
		patches["country"] = *req.Country
	}

	tx := s.db.Begin()
	tx.Where("id = ?", req.Id).Model(&User{}).Updates(patches)
	if tx.Error != nil {
		tx.Rollback()
		s.lg.Err(tx.Error).Msg("update query in UpdateUser func")

		return nil, status.Error(codes.Internal, "internal server error")
	}

	updatedUser := User{}
	tx.First(&updatedUser, "id = ?", req.Id)
	if tx.Error != nil {
		tx.Rollback()
		s.lg.Err(tx.Error).Msg("select query in UpdateUser func")

		return nil, status.Error(codes.Internal, "internal server error")
	}
	if updatedUser.ID == "" {
		tx.Rollback()

		return nil, status.Error(codes.NotFound, "id not found")
	}
	tx.Commit()

	s.notifier.Notify(&updatedUser, UpdateNotification)

	return &api.UpdateUserReply{}, nil
}

func (s *UserStore) DeleteUser(ctx context.Context, req *api.DeleteUserRequest) (*api.DeleteUserReply, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing or empty 'id' field")
	}

	userToDelete := User{}
	tx := s.db.Begin()
	tx.First(&userToDelete, "id = ?", req.Id)
	if tx.Error != nil {
		tx.Rollback()
		s.lg.Err(tx.Error).Msg("select query in DeleteUser func")

		return nil, status.Error(codes.Internal, "internal server error")
	}
	if userToDelete.ID == "" {
		tx.Rollback()

		return nil, status.Error(codes.NotFound, "id not found")
	}

	tx.Where("id = ?", req.Id).Delete(&User{})
	if tx.Error != nil {
		tx.Rollback()
		s.lg.Err(tx.Error).Msg("delete query in DeleteUser func")

		return nil, status.Error(codes.Internal, "internal server error")
	}
	tx.Commit()

	s.notifier.Notify(&userToDelete, DeleteNotification)

	return &api.DeleteUserReply{}, nil
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

func (s *UserStore) AddUser(ctx context.Context, req *api.AddUserRequest) (*api.AddUserReply, error) {
	id := uuid.New().String()

	// check for missing required fields
	if req.FirstName == "" {
		return nil, status.Error(codes.InvalidArgument, "empty or missing 'first_name' field")
	}
	if req.LastName == "" {
		return nil, status.Error(codes.InvalidArgument, "empty or missing 'last_name' field")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "empty or missing 'email' field")
	}

	newUser := &User{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Nickname:  req.Nickname,
		Password:  req.Password,
		Email:     req.Email,
		Country:   req.Country,
	}

	if tx := s.db.Create(newUser); tx.Error != nil {
		s.lg.Err(tx.Error).Msg("insert query in AddUser func")

		return nil, status.Error(codes.Internal, "internal server error")
	}

	s.notifier.Notify(newUser, AddNotification)

	return &api.AddUserReply{Id: id}, nil
}

func paginateAndFilter(page int, pageSize int, filters map[string]string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page := page
		pageSize := pageSize
		offset := (page - 1) * pageSize

		if len(filters) == 0 {
			return db.Offset(offset).Limit(pageSize)
		}

		query := ""
		var values []any

		for filterName, filterVal := range filters {
			query += "AND " + filterName + " = ?"
			values = append(values, filterVal)
		}
		query = query[3:]

		return db.Offset(offset).Where(query, values...).Limit(pageSize)
	}
}

func (s *UserStore) ListUsers(req *api.ListUsersRequest, lus api.UserStore_ListUsersServer) error {
	var users []User
	tx := s.db.Scopes(paginateAndFilter(int(req.Page), int(req.PageSize), req.Filters)).Find(&users)
	if tx.Error != nil {
		s.lg.Err(tx.Error).Msg("select query in ListUsers func")

		return status.Error(codes.Internal, "internal server error")
	}

	var err error
	for _, u := range users {
		err = lus.Send(&api.User{
			Id:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Nickname:  u.Nickname,
			Password:  u.Password,
			Email:     u.Email,
			Country:   u.Country,
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt),
		})
		if err != nil {
			s.lg.Err(err).Msg("rpc send in ListUsers func")

			return status.Error(codes.Internal, "internal server error")
		}
	}

	return nil
}
