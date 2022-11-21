package app

import "sync"

// Notifier is very simple interface that notify other systems for user changes.

type Notifier interface {
	NotifyAdd(newUser *User)
	NotifyDelete(DeletedUser *User)
	NotifyUpdate(UpdatedUser *User)
}

type notification struct {
	user   *User
	action string
}

type MockedNotifier struct {
	list []notification
	lock *sync.Mutex
}

func NewMockedNotifier() *MockedNotifier {
	return &MockedNotifier{
		list: []notification{},
		lock: &sync.Mutex{},
	}
}

func (n *MockedNotifier) NotifyAdd(newUser *User) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.list = append(n.list, notification{user: newUser, action: "add"})
}

func (n *MockedNotifier) NotifyDelete(deletedUser *User) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.list = append(n.list, notification{user: deletedUser, action: "delete"})
}

func (n *MockedNotifier) NotifyUpdate(updatedUser *User) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.list = append(n.list, notification{user: updatedUser, action: "update"})
}

func (n *MockedNotifier) GetCompiledData() string {
	n.lock.Lock()
	defer n.lock.Unlock()

	output := ""
	for _, item := range n.list {
		output += item.action + " " + item.user.FirstName + "\n"
	}

	return output
}

func (n *MockedNotifier) Reset() {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.list = []notification{}
}

func (n *MockedNotifier) ActionCallsCount(action string) int {
	n.lock.Lock()
	defer n.lock.Unlock()

	count := 0
	for _, item := range n.list {
		if item.action == action {
			count++
		}
	}

	return count
}

var _ Notifier = &MockedNotifier{}
