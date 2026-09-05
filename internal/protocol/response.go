package protocol

import (
	"errors"
	"sync"
)

const (
	InsertIdentity          = "insert.identity"
	InsertParameterIdentity = "insert.parameter.identity"
	InsertParameterDisplay  = "insert.parameter.display"
	InsertParameterValue    = "insert.parameter.value"
)

type Response struct {
	ID      uint32
	Status  string
	Value   float64
	Message string
}

type Feedback struct {
	Type         string
	Band         int
	Slot         int
	Parameter    int
	Normalized   float64
	Name         string
	Vendor       string
	Version      string
	Format       string
	DisplayValue string
}

func DecodeFeedback(frame []byte) (Feedback, error) {
	if len(frame) >= 9 && frame[0] == 0xF0 && frame[1] == 0x7D && frame[2] == 0x43 && frame[3] == 0x41 && frame[4] == Version && frame[len(frame)-1] == 0xF7 {
		feedback := Feedback{Slot: int(frame[6]), Parameter: int(frame[7])}
		parts := splitTextParts(frame[8 : len(frame)-1])
		switch frame[5] {
		case 1:
			feedback.Type = InsertIdentity
			if len(parts) > 0 {
				feedback.Name = parts[0]
			}
			if len(parts) > 1 {
				feedback.Vendor = parts[1]
			}
			if len(parts) > 2 {
				feedback.Version = parts[2]
			}
			if len(parts) > 3 {
				feedback.Format = parts[3]
			}
		case 2:
			feedback.Type = InsertParameterIdentity
			if len(parts) > 0 {
				feedback.Name = parts[0]
			}
		case 3:
			feedback.Type = InsertParameterDisplay
			if len(parts) > 0 {
				feedback.DisplayValue = parts[0]
			}
		default:
			return Feedback{}, errors.New("unknown SysEx feedback event")
		}
		return feedback, nil
	}
	if len(frame) != 3 || frame[0]&0xF0 != 0xB0 {
		return Feedback{}, errors.New("invalid feedback frame")
	}
	feedback := Feedback{Normalized: float64(frame[2]) / 127}
	cc := int(frame[1])
	switch {
	case cc == 115:
		feedback.Type = Play
	case cc == 114:
		feedback.Type = Stop
	case cc == 113:
		feedback.Type = Record
	case cc == 20:
		feedback.Type = Volume
	case cc == 21:
		feedback.Type = Pan
	case cc == 22:
		feedback.Type = Mute
	case cc == 23:
		feedback.Type = Solo
	case cc >= 40 && cc <= 55:
		feedback.Band = (cc-40)/4 + 1
		switch (cc - 40) % 4 {
		case 0:
			feedback.Type = EQGain
		case 1:
			feedback.Type = EQFreq
		case 2:
			feedback.Type = EQQ
		case 3:
			feedback.Type = EQEnable
		}
	case cc >= 60 && cc <= 67:
		feedback.Type = InsertBypass
		feedback.Slot = cc - 59
	case cc >= 80 && cc <= 87:
		feedback.Type = InsertParameterValue
		feedback.Slot = 1
		feedback.Parameter = cc - 79
	default:
		return Feedback{}, errors.New("unknown feedback control")
	}
	return feedback, nil
}

func splitTextParts(data []byte) []string {
	parts := []string{}
	start := 0
	for i, value := range data {
		if value == 0 {
			parts = append(parts, string(data[start:i]))
			start = i + 1
		}
	}
	parts = append(parts, string(data[start:]))
	return parts
}

type Matcher struct {
	mu      sync.Mutex
	waiters map[uint32]chan Response
}

func NewMatcher() *Matcher { return &Matcher{waiters: make(map[uint32]chan Response)} }

func (m *Matcher) Register(id uint32) (<-chan Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.waiters[id]; exists {
		return nil, errors.New("request ID already pending")
	}
	ch := make(chan Response, 1)
	m.waiters[id] = ch
	return ch, nil
}

func (m *Matcher) Resolve(response Response) bool {
	m.mu.Lock()
	ch, exists := m.waiters[response.ID]
	if exists {
		delete(m.waiters, response.ID)
	}
	m.mu.Unlock()
	if !exists {
		return false
	}
	ch <- response
	close(ch)
	return true
}

func (m *Matcher) Cancel(id uint32) {
	m.mu.Lock()
	ch, exists := m.waiters[id]
	if exists {
		delete(m.waiters, id)
	}
	m.mu.Unlock()
	if exists {
		close(ch)
	}
}
