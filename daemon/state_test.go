package daemon

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestStateRunStop(t *testing.T) {
	s := NewState()
	for i := 1; i < 3; i++ { // full lifecycle two times
		s.SetRunning(i + 100)
		if !s.IsRunning() {
			t.Fatal("State not running")
		}
		if s.Pid != i+100 {
			t.Fatalf("Pid %v, expected %v", s.Pid, i+100)
		}
		if s.ExitCode != 0 {
			t.Fatalf("ExitCode %v, expected 0", s.ExitCode)
		}

		stopped := make(chan struct{})
		var exit int64
		go func() {
			exitCode, err := s.WaitStop(-1 * time.Second)
			if err != nil {
				t.Fatal(err)
			}
			atomic.StoreInt64(&exit, int64(exitCode))
			close(stopped)
		}()
		s.SetStopped(i)
		if s.IsRunning() {
			t.Fatal("State is running")
		}
		if s.ExitCode != i {
			t.Fatalf("ExitCode %v, expected %v", s.ExitCode, i)
		}
		if s.Pid != 0 {
			t.Fatalf("Pid %v, expected 0", s.Pid)
		}
		select {
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Stop callback doesn't fire in 100 milliseconds")
		case <-stopped:
			t.Log("Stop callback fired")
		}
		exitCode := int(atomic.LoadInt64(&exit))
		if exitCode != i {
			t.Fatalf("ExitCode %v, expected %v", exitCode, i)
		}
		if exitCode, err := s.WaitStop(-1 * time.Second); err != nil || exitCode != i {
			t.Fatal("WaitStop returned exitCode: %v, err: %v, expected exitCode: %v, err: %v", exitCode, err, i, nil)
		}
	}
}

func TestStateTimeoutWait(t *testing.T) {
	s := NewState()
	s.SetStopped(1)
	s.SetRunning(42)
	stopped := make(chan struct{})
	go func() {
		ec, err := s.WaitStop(100 * time.Millisecond)
		if err != nil {
			t.Log(err)
		}
		if ec != -1 {
			t.Fatalf("Exit code should be -1, got %d", ec)
		}
		close(stopped)
	}()
	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop callback doesn't fire in 100 milliseconds")
	case <-stopped:
		t.Log("Stop callback fired")
	}
}

func TestStateDoubleRunDoubleStop(t *testing.T) {
	s := NewState()
	s.SetRunning(52)
	if s.Pid != 52 {
		t.Fatalf("Pid should be 52, got %d", s.Pid)
	}
	s.SetRunning(42)
	if s.Pid != 42 {
		t.Fatalf("Pid should be 42, got %d", s.Pid)
	}
	s.SetStopped(1)
	if s.ExitCode != 1 {
		t.Fatalf("ExitCde should be 1, got %d", s.ExitCode)
	}
	s.SetStopped(2)
	if s.ExitCode != 2 {
		t.Fatalf("ExitCde should be 2, got %d", s.ExitCode)
	}
}

func TestStateDoubleStopDoubleRun(t *testing.T) {
	s := NewState()
	s.SetStopped(1)
	if s.ExitCode != 1 {
		t.Fatalf("ExitCde should be 1, got %d", s.ExitCode)
	}
	s.SetStopped(2)
	if s.ExitCode != 2 {
		t.Fatalf("ExitCde should be 2, got %d", s.ExitCode)
	}
	s.SetRunning(52)
	if s.Pid != 52 {
		t.Fatalf("Pid should be 52, got %d", s.Pid)
	}
	s.SetRunning(42)
	if s.Pid != 42 {
		t.Fatalf("Pid should be 42, got %d", s.Pid)
	}
}

func TestWaitStopOnNewState(t *testing.T) {
	s := NewState()
	stopped := make(chan struct{})
	go func() {
		ec, err := s.WaitStop(-1 * time.Millisecond)
		if err != nil {
			t.Log(err)
		}
		if ec != 0 {
			t.Fatalf("Exit code should be 0, got %d", ec)
		}
		close(stopped)
	}()
	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop callback doesn't fire in 500 milliseconds")
	case <-stopped:
		t.Log("Stop callback fired")
	}
}

func TestWaitStopOnStoppedState(t *testing.T) {
	s := NewState()
	s.SetStopped(1)
	stopped := make(chan struct{})
	go func() {
		ec, err := s.WaitStop(-1 * time.Millisecond)
		if err != nil {
			t.Log(err)
		}
		if ec != 1 {
			t.Fatalf("Exit code should be 1, got %d", ec)
		}
		close(stopped)
	}()
	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop callback doesn't fire in 500 milliseconds")
	case <-stopped:
		t.Log("Stop callback fired")
	}
}
