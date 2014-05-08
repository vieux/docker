package engine

import (
	"fmt"
	"runtime"
	"testing"
)

func TestJobStatusOK(t *testing.T) {
	eng := New()
	eng.Register("return_ok", func(job *Job) Status { return StatusOK })
	err := eng.Job("return_ok").Run()
	if err != nil {
		t.Fatalf("Expected: err=%v\nReceived: err=%v", nil, err)
	}
}

func TestJobStatusErr(t *testing.T) {
	eng := New()
	eng.Register("return_err", func(job *Job) Status { return StatusErr })
	err := eng.Job("return_err").Run()
	if err == nil {
		t.Fatalf("When a job returns StatusErr, Run() should return an error")
	}
}

func TestJobStatusNotFound(t *testing.T) {
	eng := New()
	eng.Register("return_not_found", func(job *Job) Status { return StatusNotFound })
	err := eng.Job("return_not_found").Run()
	if err == nil {
		t.Fatalf("When a job returns StatusNotFound, Run() should return an error")
	}
}

func TestJobStdoutString(t *testing.T) {
	eng := New()
	// FIXME: test multiple combinations of output and status
	eng.Register("say_something_in_stdout", func(job *Job) Status {
		job.Printf("Hello world\n")
		return StatusOK
	})

	job := eng.Job("say_something_in_stdout")
	var output string
	if err := job.Stdout.AddString(&output); err != nil {
		t.Fatal(err)
	}
	if err := job.Run(); err != nil {
		t.Fatal(err)
	}
	if expectedOutput := "Hello world"; output != expectedOutput {
		t.Fatalf("Stdout last line:\nExpected: %v\nReceived: %v", expectedOutput, output)
	}
}

func TestJobStderrString(t *testing.T) {
	eng := New()
	// FIXME: test multiple combinations of output and status
	eng.Register("say_something_in_stderr", func(job *Job) Status {
		job.Errorf("Warning, something might happen\nHere it comes!\nOh no...\nSomething happened\n")
		return StatusOK
	})

	job := eng.Job("say_something_in_stderr")
	var output string
	if err := job.Stderr.AddString(&output); err != nil {
		t.Fatal(err)
	}
	if err := job.Run(); err != nil {
		t.Fatal(err)
	}
	if expectedOutput := "Something happened"; output != expectedOutput {
		t.Fatalf("Stderr last line:\nExpected: %v\nReceived: %v", expectedOutput, output)
	}
}

func TestJobAddString(t *testing.T) {
	eng := New()
	eng.Register("say_something", func(job *Job) Status {
		fmt.Fprintln(job.Stdout, "line1")
		fmt.Fprintln(job.Stdout, "line2")
		fmt.Fprintln(job.Stdout, "line3")
		return StatusOK
	})
	job := eng.Job("say_something")
	var output1, output2 string
	if err := job.Stdout.AddString(&output1); err != nil {
		t.Fatal(err)
	}
	if err := job.Stdout.AddString(&output2); err != nil {
		t.Fatal(err)
	}
	numGoroutine := runtime.NumGoroutine()
	if err := job.Run(); err != nil {
		t.Fatal(err)
	}
	if expectedOutput := "line3"; output1 != expectedOutput {
		t.Fatalf("Stderr last line:\nExpected: %v\nReceived: %v", expectedOutput, output1)
	}
	if expectedOutput := "line3"; output2 != expectedOutput {
		t.Fatalf("Stderr last line:\nExpected: %v\nReceived: %v", expectedOutput, output2)
	}
	if newNumGoroutine := runtime.NumGoroutine(); newNumGoroutine != numGoroutine {
		t.Fatalf("Goroutine leak, was %d, is now %d", numGoroutine, newNumGoroutine)
	}
}
