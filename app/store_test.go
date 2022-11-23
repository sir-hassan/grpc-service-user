package app_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/sir-hassan/grpc-service-user/api"
	"github.com/sir-hassan/grpc-service-user/app"
	"gorm.io/gorm"
)

func makeMockDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(app.User{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestUserStore_AddUser_ValidCases(t *testing.T) {
	db, err := makeMockDB()
	if err != nil {
		t.Errorf("create mock db: %v", err)
	}
	notifier := app.NewMockedNotifier()
	s := app.NewUserStore(db, notifier, zerolog.Logger{})

	tests := []struct {
		name    string
		req     *api.AddUserRequest
		wantErr bool
	}{
		{
			name: "valid_case",
			req: &api.AddUserRequest{
				FirstName: "fn1",
				LastName:  "ln1",
				Email:     "me@example.com",
			},
			wantErr: false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.AddUser(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddUser() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if _, err := uuid.Parse(got.Id); err != nil {
				t.Errorf("AddUser() replied invalid uuid: %s", got.Id)
			}
			if notifier.ActionCallsCount("add") != i+1 {
				t.Errorf("AddUser() didn't trigger notifier")
			}
		})
		if notifier.ActionCallsCount("update") != 0 {
			t.Errorf("AddUser() triggered invalid notification")
		}
		if notifier.ActionCallsCount("delete") != 0 {
			t.Errorf("AddUser() triggered invalid notification")
		}
	}
}

func TestUserStore_UpdateUser_ValidCase(t *testing.T) {
	db, err := makeMockDB()
	if err != nil {
		t.Errorf("create mock db: %v", err)
	}
	notifier := app.NewMockedNotifier()
	s := app.NewUserStore(db, notifier, zerolog.Logger{})

	var ids []string
	for i := 0; i < 10; i++ {
		u, err := s.AddUser(context.Background(), &api.AddUserRequest{
			FirstName: "user_first_name_" + strconv.Itoa(i),
			LastName:  "user_last_name_" + strconv.Itoa(i),
			Email:     "some_" + strconv.Itoa(i) + "@example.com",
		})
		if err != nil {
			t.Errorf("unexpected error on call add user: %v", err)
		}
		ids = append(ids, u.Id)
	}

	// update first user
	str := "updated_first_name"
	_, err = s.UpdateUser(context.Background(), &api.UpdateUserRequest{
		Id:        ids[0],
		FirstName: &str,
	})
	if err != nil {
		t.Errorf("unexpected error on call update user: %v", err)
	}

	if notifier.ActionCallsCount("add") != 10 {
		t.Errorf("AddUser() triggered invalid notification")
	}
	if notifier.ActionCallsCount("update") != 1 {
		t.Errorf("AddUser() triggered invalid notification")
	}
	if notifier.ActionCallsCount("delete") != 0 {
		t.Errorf("AddUser() triggered invalid notification")
	}
}

func TestUserStore_DeleteUser_ValidCase(t *testing.T) {
	db, err := makeMockDB()
	if err != nil {
		t.Errorf("create mock db: %v", err)
	}
	notifier := app.NewMockedNotifier()
	s := app.NewUserStore(db, notifier, zerolog.Logger{})

	var ids []string
	for i := 0; i < 10; i++ {
		u, err := s.AddUser(context.Background(), &api.AddUserRequest{
			FirstName: "user_first_name_" + strconv.Itoa(i),
			LastName:  "user_last_name_" + strconv.Itoa(i),
			Email:     "some_" + strconv.Itoa(i) + "@example.com",
		})
		if err != nil {
			t.Errorf("unexpected error on call add user: %v", err)
		}
		ids = append(ids, u.Id)
	}

	// delete first user
	_, err = s.DeleteUser(context.Background(), &api.DeleteUserRequest{
		Id: ids[0],
	})
	if err != nil {
		t.Errorf("unexpected error on call delete user: %v", err)
	}

	if notifier.ActionCallsCount("add") != 10 {
		t.Errorf("AddUser() triggered invalid notification")
	}
	if notifier.ActionCallsCount("update") != 0 {
		t.Errorf("AddUser() triggered invalid notification")
	}
	if notifier.ActionCallsCount("delete") != 1 {
		t.Errorf("AddUser() triggered invalid notification")
	}
}

func TestUserStore_DeleteUser_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantError string
	}{
		{
			name:      "invalid id",
			id:        "some_uuid",
			wantError: "code = NotFound",
		},
		{
			name:      "empty id",
			id:        "",
			wantError: "code = InvalidArgument",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := makeMockDB()
			if err != nil {
				t.Errorf("create mock db: %v", err)
			}
			notifier := app.NewMockedNotifier()
			s := app.NewUserStore(db, notifier, zerolog.Logger{})

			reply, err := s.DeleteUser(context.Background(), &api.DeleteUserRequest{
				Id: tt.id,
			})
			if reply != nil {
				t.Errorf("expect nil reply, got: %v", reply)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("unexpected error, want: %s, got: %s", tt.wantError, err.Error())
			}

			if notifier.ActionCallsCount("add") != 0 {
				t.Errorf("AddUser() triggered invalid notification")
			}
			if notifier.ActionCallsCount("update") != 0 {
				t.Errorf("AddUser() triggered invalid notification")
			}
			if notifier.ActionCallsCount("delete") != 0 {
				t.Errorf("AddUser() triggered invalid notification")
			}
		})
	}
}

func TestUserStore_UpdateUser_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantError string
	}{
		{
			name:      "invalid id",
			id:        "some_uuid",
			wantError: "code = NotFound",
		},
		{
			name:      "empty id",
			id:        "",
			wantError: "code = InvalidArgument",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := makeMockDB()
			if err != nil {
				t.Errorf("create mock db: %v", err)
			}
			notifier := app.NewMockedNotifier()
			s := app.NewUserStore(db, notifier, zerolog.Logger{})

			someStr := "someString"
			reply, err := s.UpdateUser(context.Background(), &api.UpdateUserRequest{
				Id:        tt.id,
				FirstName: &someStr,
			})
			if reply != nil {
				t.Errorf("expect nil reply, got: %v", reply)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("unexpected error, want: %s, got: %s", tt.wantError, err.Error())
			}

			if notifier.ActionCallsCount("add") != 0 {
				t.Errorf("AddUser() triggered invalid notification")
			}
			if notifier.ActionCallsCount("update") != 0 {
				t.Errorf("AddUser() triggered invalid notification")
			}
			if notifier.ActionCallsCount("delete") != 0 {
				t.Errorf("AddUser() triggered invalid notification")
			}
		})
	}
}
