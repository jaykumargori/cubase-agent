package tests

import (
	"github.com/jaykumargori/cubase-agent/internal/protocol"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	b, e := protocol.Encode(protocol.Command{ID: 42, Type: protocol.Play})
	if e != nil {
		t.Fatal(e)
	}
	c, e := protocol.Decode(b)
	if e != nil || c.Type != protocol.Play {
		t.Fatalf("%+v %v", c, e)
	}
}

func TestRecordRoundTrip(t *testing.T) {
	b, err := protocol.Encode(protocol.Command{ID: 43, Type: protocol.Record})
	if err != nil {
		t.Fatal(err)
	}
	if b[1] != 113 {
		t.Fatalf("record CC = %d, want 113", b[1])
	}
	c, err := protocol.Decode(b)
	if err != nil || c.Type != protocol.Record {
		t.Fatalf("%+v %v", c, err)
	}
}

func TestInsertBypassEncoding(t *testing.T) {
	b, err := protocol.EncodeInsertBypass(8, 127)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 3 || b[0] != 0xB0 || b[1] != 67 || b[2] != 127 {
		t.Fatalf("unexpected insert frame: %v", b)
	}
	if _, err = protocol.EncodeInsertBypass(9, 0); err == nil {
		t.Fatal("expected slot range error")
	}
}

func TestPluginParameterEncoding(t *testing.T) {
	b, err := protocol.EncodePluginParameter(1, 64)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 3 || b[0] != 0xB0 || b[1] != 80 || b[2] != 64 {
		t.Fatalf("unexpected plugin parameter frame: %v", b)
	}
	if _, err = protocol.EncodePluginParameter(9, 0); err == nil {
		t.Fatal("expected parameter range error")
	}
}

func TestFeedbackDecoding(t *testing.T) {
	feedback, err := protocol.DecodeFeedback([]byte{0xB0, 45, 64})
	if err != nil {
		t.Fatal(err)
	}
	if feedback.Type != protocol.EQFreq || feedback.Band != 2 {
		t.Fatalf("feedback: %+v", feedback)
	}
	feedback, err = protocol.DecodeFeedback([]byte{0xB0, 83, 127})
	if err != nil {
		t.Fatal(err)
	}
	if feedback.Slot != 1 || feedback.Parameter != 4 || feedback.Normalized != 1 {
		t.Fatalf("feedback: %+v", feedback)
	}
}

func TestSysExIdentityFeedback(t *testing.T) {
	frame := append([]byte{0xF0, 0x7D, 0x43, 0x41, protocol.Version, 1, 2, 0}, []byte("Compressor\x00Steinberg\x001.0\x00VST3")...)
	frame = append(frame, 0xF7)
	feedback, err := protocol.DecodeFeedback(frame)
	if err != nil {
		t.Fatal(err)
	}
	if feedback.Type != "insert.identity" || feedback.Slot != 2 || feedback.Name != "Compressor" || feedback.Vendor != "Steinberg" {
		t.Fatalf("feedback: %+v", feedback)
	}
}
