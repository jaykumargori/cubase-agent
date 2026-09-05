package cubase

import (
	"sort"

	"github.com/jaykumargori/cubase-agent/internal/protocol"
)

// InsertSnapshot is the event-derived view of an insert slot. MIDI Remote
// publishes this data whenever Cubase refreshes the selected channel's inserts.
type InsertSnapshot struct {
	Insert
	Vendor  string
	Version string
	Format  string
}

// BuildInsertSnapshots turns asynchronous MIDI Remote feedback into stable,
// slot-sorted insert data for the CLI and future MCP layer.
func BuildInsertSnapshots(events []protocol.Feedback) []InsertSnapshot {
	bySlot := make(map[int]*InsertSnapshot)
	ensure := func(slot int) *InsertSnapshot {
		if slot < 1 || slot > 8 {
			return nil
		}
		if bySlot[slot] == nil {
			bySlot[slot] = &InsertSnapshot{Insert: Insert{Slot: slot}}
		}
		return bySlot[slot]
	}
	parameters := make(map[int]map[int]*PluginParameter)
	for _, event := range events {
		snapshot := ensure(event.Slot)
		if snapshot == nil {
			continue
		}
		switch event.Type {
		case protocol.InsertIdentity:
			snapshot.Name = event.Name
			snapshot.Vendor = event.Vendor
			snapshot.Version = event.Version
			snapshot.Format = event.Format
		case protocol.InsertBypass:
			snapshot.Bypassed = event.Normalized >= 0.5
		case protocol.InsertParameterIdentity, protocol.InsertParameterDisplay, protocol.InsertParameterValue:
			if event.Parameter < 1 || event.Parameter > 8 {
				continue
			}
			if parameters[event.Slot] == nil {
				parameters[event.Slot] = make(map[int]*PluginParameter)
			}
			parameter := parameters[event.Slot][event.Parameter]
			if parameter == nil {
				parameter = &PluginParameter{ID: stringParameterID(event.Parameter)}
				parameters[event.Slot][event.Parameter] = parameter
			}
			switch event.Type {
			case protocol.InsertParameterIdentity:
				parameter.Name = event.Name
			case protocol.InsertParameterDisplay:
				parameter.DisplayValue = event.DisplayValue
			case protocol.InsertParameterValue:
				parameter.NormalizedValue = event.Normalized
			}
		}
	}

	slots := make([]int, 0, len(bySlot))
	for slot := range bySlot {
		slots = append(slots, slot)
	}
	sort.Ints(slots)
	result := make([]InsertSnapshot, 0, len(slots))
	for _, slot := range slots {
		snapshot := bySlot[slot]
		parameterNumbers := make([]int, 0, len(parameters[slot]))
		for parameter := range parameters[slot] {
			parameterNumbers = append(parameterNumbers, parameter)
		}
		sort.Ints(parameterNumbers)
		for _, parameter := range parameterNumbers {
			snapshot.Parameters = append(snapshot.Parameters, *parameters[slot][parameter])
		}
		result = append(result, *snapshot)
	}
	return result
}

func stringParameterID(parameter int) string {
	return string(rune('0' + parameter))
}
