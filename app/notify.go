package app

import "sync"

// Notifier is very simple interface that notify other systems for user changes.

type Notifier interface{}

type notification struct {
	user   *User
	action string
}

type MockedNotifier struct {
	list []notification
	lock *sync.Mutex
}

func NewMockedNotifier() *MockedNotifier {
	return &MockedNotifier{}
}

var _ Notifier = &MockedNotifier{}
