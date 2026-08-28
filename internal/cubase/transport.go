package cubase

import (
	"fmt"
	"github.com/jaykumargori/cubase-agent/internal/protocol"
)

type Sender interface{ Send([]byte) error }

type Bridge struct {
	MIDI Sender
	next uint32
}

func (b *Bridge) Play() error   { return b.Send(protocol.Play) }
func (b *Bridge) Stop() error   { return b.Send(protocol.Stop) }
func (b *Bridge) Record() error { return b.Send(protocol.Record) }

func (b *Bridge) Send(kind string) error {
	return b.SendValue(kind, 127)
}
func (b *Bridge) SendValue(kind string, value byte) error {
	b.next++
	frame, err := protocol.Encode(protocol.Command{ID: b.next, Type: kind})
	if err != nil {
		return err
	}
	frame[2] = value
	if err = b.MIDI.Send(frame); err != nil {
		return fmt.Errorf("%s: %w", kind, err)
	}
	return nil
}
func (b *Bridge) SendEQ(kind string, band int, value byte) error {
	frame, err := protocol.EncodeEQ(kind, band, value)
	if err != nil {
		return err
	}
	if err = b.MIDI.Send(frame); err != nil {
		return err
	}
	return nil
}
func (b *Bridge) SendInsertBypass(slot int, value byte) error {
	frame, err := protocol.EncodeInsertBypass(slot, value)
	if err != nil {
		return err
	}
	return b.MIDI.Send(frame)
}
func (b *Bridge) SendPluginParameter(parameter int, value byte) error {
	frame, err := protocol.EncodePluginParameter(parameter, value)
	if err != nil {
		return err
	}
	return b.MIDI.Send(frame)
}
