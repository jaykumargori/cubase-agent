package cubase

import (
	"fmt"
	"math"

	"github.com/jaykumargori/cubase-agent/internal/protocol"
)

func normalized(value, min, max float64) (byte, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < min || value > max {
		return 0, fmt.Errorf("value %.4g outside %.4g..%.4g", value, min, max)
	}
	return byte(math.Round((value - min) / (max - min) * 127)), nil
}

func boolValue(enabled bool) byte {
	if enabled {
		return 127
	}
	return 0
}

func (b *Bridge) SetSelectedTrackVolume(value float64) error {
	v, err := normalized(value, 0, 1)
	if err != nil {
		return err
	}
	return b.SendValue(protocol.Volume, v)
}

func (b *Bridge) SetSelectedTrackPan(value float64) error {
	v, err := normalized(value, -1, 1)
	if err != nil {
		return err
	}
	return b.SendValue(protocol.Pan, v)
}

func (b *Bridge) SetSelectedTrackMute(enabled bool) error {
	return b.SendValue(protocol.Mute, boolValue(enabled))
}
func (b *Bridge) SetSelectedTrackSolo(enabled bool) error {
	return b.SendValue(protocol.Solo, boolValue(enabled))
}
