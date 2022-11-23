package app

import (
	"fmt"
	"sync"
)

// Notifier is very simple interface that notify other systems for user changes.
// The implementations should consider being asynchronous, this is why there are no errors returned here.
// Errors should be handled and logged by the concrete implementations.

type NotificationType int

const (
	AddNotification NotificationType = iota
	DeleteNotification
	UpdateNotification
)

type Notifier interface {
	Notify(user *User, typ NotificationType)
}

type MockedNotifier struct {
	actionsList []string
	lock        *sync.Mutex
}

func NewMockedNotifier() *MockedNotifier {
	return &MockedNotifier{
		actionsList: []string{},
		lock:        &sync.Mutex{},
	}
}

func (n *MockedNotifier) Notify(user *User, typ NotificationType) {
	n.lock.Lock()
	defer n.lock.Unlock()

	switch typ {
	case UpdateNotification:
		n.actionsList = append(n.actionsList, "update")
	case DeleteNotification:
		n.actionsList = append(n.actionsList, "delete")
	case AddNotification:
		n.actionsList = append(n.actionsList, "add")
	default:
		panic(fmt.Sprintf("logic error, unexpected typ: %v in Notify()\n", typ))
	}
}

func (n *MockedNotifier) Reset() {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.actionsList = []string{}
}

func (n *MockedNotifier) ActionCallsCount(action string) int {
	n.lock.Lock()
	defer n.lock.Unlock()

	count := 0
	for _, item := range n.actionsList {
		if item == action {
			count++
		}
	}

	return count
}

var _ Notifier = &MockedNotifier{}
