package app_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/sir-hassan/grpc-service-user/app"
)

func TestHTTPNotifier_Notify(t *testing.T) {
	lock := &sync.Mutex{}
	var webHookCalls []string

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		data, _ := io.ReadAll(r.Body)
		webHookCalls = append(webHookCalls, r.URL.Path+" -> "+string(data))
		_, _ = w.Write([]byte("ok"))
	}))
	defer svr.Close()

	notifier := app.NewHTTPNotifier(zerolog.Logger{}, http.DefaultClient, []string{svr.URL}, 10)
	cancelNotifierChan := make(chan any)
	doneNotifierChan := notifier.Start(cancelNotifierChan)

	notifier.Notify(&app.User{
		ID: "111",
	}, app.AddNotification)

	notifier.Notify(&app.User{
		ID: "222",
	}, app.UpdateNotification)

	notifier.Notify(&app.User{
		ID: "333",
	}, app.DeleteNotification)

	time.Sleep(time.Millisecond * 100)

	close(cancelNotifierChan)
	<-doneNotifierChan

	expectHTTPCalls := []string{
		//nolint
		`/add -> {"ID":"111","FirstName":"","LastName":"","Nickname":"","Password":"","Email":"","Country":"","CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z"}`,
		//nolint
		`/update -> {"ID":"222","FirstName":"","LastName":"","Nickname":"","Password":"","Email":"","Country":"","CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z"}`,
		//nolint
		`/delete -> {"ID":"333","FirstName":"","LastName":"","Nickname":"","Password":"","Email":"","Country":"","CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z"}`,
	}

	if !reflect.DeepEqual(webHookCalls, expectHTTPCalls) {
		t.Errorf("Unexpected webhook calls = %v, want %v", webHookCalls, expectHTTPCalls)
	}
}
