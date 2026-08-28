package cubase

import (
	"fmt"
	"math"

	"github.com/jaykumargori/cubase-agent/internal/protocol"
)

type EQBand struct {
	Enabled   bool
	Frequency float64
	GainDB    float64
	Q         float64
}

func (b *Bridge) SetEQGain(band int, db float64) error {
	v, err := normalized(db, -24, 24)
	if err != nil {
		return fmt.Errorf("EQ gain: %w", err)
	}
	return b.SendEQ(protocol.EQGain, band, v)
}

func (b *Bridge) SetEQFrequency(band int, hz float64) error {
	if math.IsNaN(hz) || math.IsInf(hz, 0) || hz < 20 || hz > 20000 {
		return fmt.Errorf("EQ frequency must be 20..20000 Hz")
	}
	v := byte(math.Round(math.Log(hz/20) / math.Log(20000/20) * 127))
	return b.SendEQ(protocol.EQFreq, band, v)
}

func (b *Bridge) SetEQQ(band int, q float64) error {
	if math.IsNaN(q) || math.IsInf(q, 0) || q < 0.1 || q > 20 {
		return fmt.Errorf("EQ Q must be 0.1..20")
	}
	v := byte(math.Round(math.Log(q/0.1) / math.Log(20/0.1) * 127))
	return b.SendEQ(protocol.EQQ, band, v)
}

func (b *Bridge) EnableEQBand(band int, enabled bool) error {
	return b.SendEQ(protocol.EQEnable, band, boolValue(enabled))
}
