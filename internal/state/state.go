// Package state maintains the event-derived Cubase readback snapshot.
package state

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/jaykumargori/cubase-agent/internal/cubase"
	"github.com/jaykumargori/cubase-agent/internal/protocol"
)

type Connection struct {
	OutputPort        string
	Connected         bool
	ReadbackAvailable bool
	CheckedAt         time.Time
}

type Mixer struct {
	Volume float64 `json:"volume"`
	Pan    float64 `json:"pan"`
	Mute   bool    `json:"mute"`
	Solo   bool    `json:"solo"`
}
type EQBand struct {
	Band      int     `json:"band"`
	GainDB    float64 `json:"gainDb"`
	Frequency float64 `json:"frequencyHz"`
	Q         float64 `json:"q"`
	Enabled   bool    `json:"enabled"`
}
type Snapshot struct {
	Mixer              Mixer                   `json:"mixer"`
	EQ                 []EQBand                `json:"eq"`
	Inserts            []cubase.InsertSnapshot `json:"inserts"`
	LastTransportEvent string                  `json:"lastTransportEvent,omitempty"`
	ReadbackAvailable  bool                    `json:"readbackAvailable"`
	LastFeedbackAt     time.Time               `json:"lastFeedbackAt,omitempty"`
}

type Store struct {
	mu       sync.RWMutex
	snapshot Snapshot
	events   []protocol.Feedback
}

func New() *Store {
	bands := make([]EQBand, 4)
	for i := range bands {
		bands[i].Band = i + 1
	}
	return &Store{snapshot: Snapshot{EQ: bands}}
}

func (s *Store) Apply(event protocol.Feedback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.ReadbackAvailable = true
	s.snapshot.LastFeedbackAt = time.Now().UTC()
	switch event.Type {
	case protocol.Play, protocol.Stop, protocol.Record:
		s.snapshot.LastTransportEvent = event.Type
	case protocol.Volume:
		s.snapshot.Mixer.Volume = event.Normalized
	case protocol.Pan:
		s.snapshot.Mixer.Pan = event.Normalized*2 - 1
	case protocol.Mute:
		s.snapshot.Mixer.Mute = event.Normalized >= 0.5
	case protocol.Solo:
		s.snapshot.Mixer.Solo = event.Normalized >= 0.5
	case protocol.EQGain:
		s.band(event.Band).GainDB = event.Normalized*48 - 24
	case protocol.EQFreq:
		s.band(event.Band).Frequency = 20 * math.Pow(1000, event.Normalized)
	case protocol.EQQ:
		s.band(event.Band).Q = 0.1 * math.Pow(200, event.Normalized)
	case protocol.EQEnable:
		s.band(event.Band).Enabled = event.Normalized >= 0.5
	}
	s.events = append(s.events, event)
	if len(s.events) > 256 {
		s.events = append([]protocol.Feedback(nil), s.events[len(s.events)-256:]...)
	}
	s.snapshot.Inserts = cubase.BuildInsertSnapshots(s.events)
}

func (s *Store) band(number int) *EQBand {
	if number < 1 || number > len(s.snapshot.EQ) {
		return &EQBand{}
	}
	return &s.snapshot.EQ[number-1]
}
func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := s.snapshot
	result.EQ = append([]EQBand(nil), s.snapshot.EQ...)
	result.Inserts = append([]cubase.InsertSnapshot(nil), s.snapshot.Inserts...)
	for i := range result.Inserts {
		result.Inserts[i].Parameters = append([]cubase.PluginParameter(nil), s.snapshot.Inserts[i].Parameters...)
	}
	return result
}

type Receiver interface {
	Receive(time.Duration) ([]byte, error)
}

func Pump(ctx context.Context, receiver Receiver, store *Store) {
	for ctx.Err() == nil {
		frame, err := receiver.Receive(250 * time.Millisecond)
		if err != nil {
			continue
		}
		feedback, err := protocol.DecodeFeedback(frame)
		if err == nil {
			store.Apply(feedback)
		}
	}
}
