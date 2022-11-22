package app

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

type HTTPNotifier struct {
	hook   string
	lg     zerolog.Logger
	client *http.Client
}

func NewHTTPNotifier(lg zerolog.Logger) *HTTPNotifier {
	return &HTTPNotifier{
		// http.DefaultClient should handle cashing and connection pooling out of the box.
		client: http.DefaultClient,
		lg:     lg,
	}
}

func (n *HTTPNotifier) notify(user *User, action string) {
	n.hook = "https://webhook.site/49525789-a6ae-4f41-8518-4c2d3ae8f4c3"
	url := n.hook + "/" + action
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

func (n *HTTPNotifier) NotifyAdd(newUser *User) {
	n.notify(newUser, "add")
}

func (n *HTTPNotifier) NotifyDelete(deletedUser *User) {
	n.notify(deletedUser, "delete")
}

func (n *HTTPNotifier) NotifyUpdate(updatedUser *User) {
	n.notify(updatedUser, "update")
}

var _ Notifier = &HTTPNotifier{}
