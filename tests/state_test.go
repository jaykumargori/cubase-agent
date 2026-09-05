package tests

import (
	"math"
	"testing"

	"github.com/jaykumargori/cubase-agent/internal/protocol"
	"github.com/jaykumargori/cubase-agent/internal/state"
)

func TestStoreAppliesMixerEQAndInsertFeedback(t *testing.T) {
	store := state.New()
	store.Apply(protocol.Feedback{Type: protocol.Volume, Normalized: 0.7})
	store.Apply(protocol.Feedback{Type: protocol.Pan, Normalized: 0.25})
	store.Apply(protocol.Feedback{Type: protocol.EQGain, Band: 2, Normalized: 0.75})
	store.Apply(protocol.Feedback{Type: protocol.InsertIdentity, Slot: 1, Name: "Compressor"})
	store.Apply(protocol.Feedback{Type: protocol.InsertParameterIdentity, Slot: 1, Parameter: 1, Name: "Threshold"})
	snapshot := store.Snapshot()
	if !snapshot.ReadbackAvailable || snapshot.Mixer.Volume != 0.7 || snapshot.Mixer.Pan != -0.5 {
		t.Fatalf("unexpected mixer snapshot: %+v", snapshot.Mixer)
	}
	if math.Abs(snapshot.EQ[1].GainDB-12) > 0.001 {
		t.Fatalf("EQ gain = %f", snapshot.EQ[1].GainDB)
	}
	if len(snapshot.Inserts) != 1 || snapshot.Inserts[0].Name != "Compressor" || snapshot.Inserts[0].Parameters[0].Name != "Threshold" {
		t.Fatalf("unexpected inserts: %+v", snapshot.Inserts)
	}
}
