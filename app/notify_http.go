package app

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

type queueMessage struct {
	webHook string
	user    *User
	typ     NotificationType
}

// HTTPNotifier implements Notifier in an asynchronous manner. HTTPNotifier appends notifications to be sent in a
// channel and a goroutine (spawned by Start()) consumes that channel and fires the http requests. This is a very simple
// FIFO queueing solutions.
type HTTPNotifier struct {
	lg     zerolog.Logger
	client *http.Client

	webHooks []string

	// Very simple fifo queue to implement an asynchronous Notifier.
	queue chan queueMessage
}

func NewHTTPNotifier(lg zerolog.Logger, httpClient *http.Client, webHooks []string, queueSize int) *HTTPNotifier {
	return &HTTPNotifier{
		lg:       lg,
		client:   httpClient,
		webHooks: webHooks,
		queue:    make(chan queueMessage, queueSize),
	}
}

func (n *HTTPNotifier) Start(cancelChan chan any) chan any {
	doneChan := make(chan any)
	go func() {
	loop:
		for {
			select {
			case msg := <-n.queue:
				n.notify(msg.webHook, msg.user, msg.typ)
			case <-cancelChan:
				break loop
			}
		}
		n.lg.Info().Msg("http notifier stopped")
		close(doneChan)
	}()

	n.lg.Info().Msg("http notifier started")

	return doneChan
}

func (n *HTTPNotifier) notify(webhook string, user *User, typ NotificationType) {
	var action string
	switch typ {
	case UpdateNotification:
		action = "update"
	case DeleteNotification:
		action = "delete"
	case AddNotification:
		action = "add"
	default:
		n.lg.Fatal().Int("typ", int(typ)).Msg("logic error, unexpected typ value")
	}

	url := webhook + "/" + action
	jsonStr, err := json.Marshal(user)
	if err != nil {
		n.lg.Err(err).Str("url", url).Msg("marshaling post data to webhook ")

		return
	}

	//nolint
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		n.lg.Err(err).Str("url", url).Msg("firing post request to webhook")

		return
	}
	if resp.StatusCode != http.StatusOK {
		n.lg.Err(err).Str("url", url).Str("status", resp.Status).Msg("none ok post request to webhook")

		return
	}
	n.lg.Info().Str("url", url).Msg("post request to webhook successful")
}

func (n *HTTPNotifier) Notify(user *User, typ NotificationType) {
	for _, webHook := range n.webHooks {
		n.queue <- queueMessage{
			webHook: webHook,
			user:    user,
			typ:     typ,
		}
	}
}

var _ Notifier = &HTTPNotifier{}
