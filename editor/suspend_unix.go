// +build !windows,!linux

package editor

import "syscall"

func suspend(e *Editor) error {
	if err := e.ui.Close(); err != nil {
		return err
	}
	pid := syscall.Getpid()
	if err := syscall.Kill(pid, syscall.SIGSTOP); err != nil {
		return err
	}
	if err := e.ui.Init(e.uiEventCh); err != nil {
		return err
	}
	if err := e.redraw(); err != nil {
		return err
	}
	go e.ui.Run(defaultKeyManagers())
	return nil
}
