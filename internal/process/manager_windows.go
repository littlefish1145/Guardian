// +build windows

package process

func (m *Manager) reapZombies() {
	// Windows handles process cleanup automatically
	// No need to manually reap zombie processes
}
