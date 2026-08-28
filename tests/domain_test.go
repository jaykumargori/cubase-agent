package tests

import (
	"testing"
	"time"

	"github.com/jaykumargori/cubase-agent/internal/cubase"
	"github.com/jaykumargori/cubase-agent/internal/protocol"
)

type captureSender struct{ frame []byte }

func (s *captureSender) Send(frame []byte) error {
	s.frame = append(s.frame[:0], frame...)
	return nil
}

func TestDomainNormalization(t *testing.T) {
	sender := &captureSender{}
	bridge := cubase.Bridge{MIDI: sender}
	if err := bridge.SetSelectedTrackVolume(0.5); err != nil {
		t.Fatal(err)
	}
	if sender.frame[1] != 20 || sender.frame[2] != 64 {
		t.Fatalf("volume frame: %v", sender.frame)
	}
	if err := bridge.SetSelectedTrackPan(2); err == nil {
		t.Fatal("expected pan range error")
	}
	if err := bridge.SetEQFrequency(2, 1000); err != nil {
		t.Fatal(err)
	}
	if sender.frame[1] != 45 {
		t.Fatalf("EQ frequency frame: %v", sender.frame)
	}
}

func TestResponseMatcher(t *testing.T) {
	matcher := protocol.NewMatcher()
	waiter, err := matcher.Register(42)
	if err != nil {
		t.Fatal(err)
	}
	if !matcher.Resolve(protocol.Response{ID: 42, Status: "ok", Value: 0.5}) {
		t.Fatal("response was not matched")
	}
	select {
	case response := <-waiter:
		if response.Status != "ok" || response.Value != 0.5 {
			t.Fatalf("response: %+v", response)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("response match timed out")
	}
	if matcher.Resolve(protocol.Response{ID: 99}) {
		t.Fatal("unexpected unmatched response")
	}
}

func TestResponseMatcherCancellation(t *testing.T) {
	matcher := protocol.NewMatcher()
	waiter, err := matcher.Register(7)
	if err != nil {
		t.Fatal(err)
	}
	matcher.Cancel(7)
	select {
	case _, open := <-waiter:
		if open {
			t.Fatal("cancelled waiter remained open")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("cancel did not close waiter")
	}
}
