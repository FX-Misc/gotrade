package gotrade

import (
	"time"
)

type Manager struct {
	Channel chan bool
	Status  bool
}

func NewManager() *Manager {
	manager := &Manager{}
	manager.Status = true
	manager.Channel = make(chan bool)
	go func() {
		for signal := range manager.Channel {
			if signal {
				manager.Status = true
			} else {
				manager.Status = false
			}
		}
	}()
	return manager
}

func (m *Manager) Listen() {
	for !m.Status {
		time.Sleep(1 * time.Second)
	}
}

func (m *Manager) GetStatus() bool {
	return m.Status
}

func (m *Manager) Pause() {
	m.Channel <- false
}

func (m *Manager) Start() {
	m.Channel <- true
}
