package tests

import (
	"testing"

	"github.com/jaykumargori/cubase-agent/internal/cubase"
	"github.com/jaykumargori/cubase-agent/internal/protocol"
)

func TestBuildInsertSnapshots(t *testing.T) {
	snapshots := cubase.BuildInsertSnapshots([]protocol.Feedback{
		{Type: protocol.InsertIdentity, Slot: 2, Name: "Compressor", Vendor: "Steinberg", Version: "1.0", Format: "VST3"},
		{Type: protocol.InsertBypass, Slot: 2, Normalized: 1},
		{Type: protocol.InsertIdentity, Slot: 1, Name: "EQ"},
		{Type: protocol.InsertParameterDisplay, Slot: 1, Parameter: 2, DisplayValue: "1.2 dB"},
		{Type: protocol.InsertParameterIdentity, Slot: 1, Parameter: 2, Name: "Gain"},
		{Type: protocol.InsertParameterValue, Slot: 1, Parameter: 2, Normalized: 0.75},
	})
	if len(snapshots) != 2 || snapshots[0].Slot != 1 || snapshots[1].Slot != 2 {
		t.Fatalf("unexpected snapshots: %+v", snapshots)
	}
	parameter := snapshots[0].Parameters[0]
	if parameter.ID != "2" || parameter.Name != "Gain" || parameter.DisplayValue != "1.2 dB" || parameter.NormalizedValue != 0.75 {
		t.Fatalf("unexpected parameter: %+v", parameter)
	}
	if !snapshots[1].Bypassed || snapshots[1].Vendor != "Steinberg" {
		t.Fatalf("unexpected insert: %+v", snapshots[1])
	}
}
