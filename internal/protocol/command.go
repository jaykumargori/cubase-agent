package protocol

import "errors"

const (
	Magic        byte = 0xCA
	Version      byte = 1
	Play              = "transport.play"
	Stop              = "transport.stop"
	Record            = "transport.record"
	Volume            = "track.volume.set"
	Pan               = "track.pan.set"
	Mute              = "track.mute.set"
	Solo              = "track.solo.set"
	EQGain            = "eq.band.gain.set"
	EQFreq            = "eq.band.frequency.set"
	EQQ               = "eq.band.q.set"
	EQEnable          = "eq.band.enable"
	InsertBypass      = "insert.bypass.set"
)

type Command struct {
	ID     uint32
	Type   string
	Target string
	Params map[string]float64
}

func EncodeEQ(kind string, band int, value byte) ([]byte, error) {
	if band < 1 || band > 4 {
		return nil, errors.New("EQ band must be 1..4")
	}
	base := byte(40 + (band-1)*4)
	switch kind {
	case EQGain:
	case EQFreq:
		base++
	case EQQ:
		base += 2
	case EQEnable:
		base += 3
	default:
		return nil, errors.New("unsupported EQ command")
	}
	return []byte{0xB0, base, value}, nil
}

func EncodeInsertBypass(slot int, value byte) ([]byte, error) {
	if slot < 1 || slot > 8 {
		return nil, errors.New("insert slot must be 1..8")
	}
	return []byte{0xB0, byte(60 + slot - 1), value}, nil
}

func EncodePluginParameter(parameter int, value byte) ([]byte, error) {
	if parameter < 1 || parameter > 8 {
		return nil, errors.New("plugin parameter must be 1..8")
	}
	return []byte{0xB0, byte(79 + parameter), value}, nil
}

// Encode uses MIDI CC on channel 1. CC 115=play, CC 114=stop.
func Encode(c Command) ([]byte, error) {
	var cc byte
	switch c.Type {
	case Play:
		cc = 115
	case Stop:
		cc = 114
	case Record:
		cc = 113
	case Volume:
		cc = 20
	case Pan:
		cc = 21
	case Mute:
		cc = 22
	case Solo:
		cc = 23
	default:
		return nil, errors.New("unsupported command: " + c.Type)
	}
	return []byte{0xB0, cc, 127}, nil
}

func Decode(b []byte) (Command, error) {
	if len(b) != 3 || b[0]&0xF0 != 0xB0 {
		return Command{}, errors.New("invalid frame")
	}
	c := Command{}
	switch b[1] {
	case 115:
		c.Type = Play
	case 114:
		c.Type = Stop
	case 113:
		c.Type = Record
	case 20:
		c.Type = Volume
	case 21:
		c.Type = Pan
	case 22:
		c.Type = Mute
	case 23:
		c.Type = Solo
	default:
		return Command{}, errors.New("unknown command")
	}
	return c, nil
}
