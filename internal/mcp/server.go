// Package mcp exposes the reversible Cubase bridge controls over MCP stdio.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/jaykumargori/cubase-agent/internal/cubase"
	"github.com/jaykumargori/cubase-agent/internal/state"
)

const protocolVersion = "2025-03-26"

type Server struct {
	Controller cubase.Controller
	State      *state.Store
	In         io.Reader
	Out        io.Writer
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type toolCall struct {
	Name      string                     `json:"name"`
	Arguments map[string]json.RawMessage `json:"arguments"`
}

func (s Server) Run() error {
	scanner := bufio.NewScanner(s.In)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for scanner.Scan() {
		var request request
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			if err := s.writeError(nil, -32700, "parse error"); err != nil {
				return err
			}
			continue
		}
		if request.JSONRPC != "2.0" || request.Method == "" {
			if err := s.writeError(request.ID, -32600, "invalid request"); err != nil {
				return err
			}
			continue
		}
		if len(request.ID) == 0 || string(request.ID) == "null" {
			continue // notifications intentionally do not receive a response.
		}
		if err := s.handle(request); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s Server) handle(request request) error {
	switch request.Method {
	case "initialize":
		return s.writeResult(request.ID, map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]string{"name": "cubase-agent", "version": "0.1.0"},
			"instructions":    "Controls only reversible transport, mixer, EQ, and insert settings. It never saves, closes, deletes, renders, or exports Cubase projects.",
		})
	case "ping":
		return s.writeResult(request.ID, map[string]any{})
	case "tools/list":
		return s.writeResult(request.ID, map[string]any{"tools": tools()})
	case "tools/call":
		var call toolCall
		if err := json.Unmarshal(request.Params, &call); err != nil {
			return s.writeError(request.ID, -32602, "invalid tools/call parameters")
		}
		if result, handled := s.read(call); handled {
			return s.writeResult(request.ID, result)
		}
		if err := s.call(call); err != nil {
			return s.writeResult(request.ID, toolResult(err.Error(), true))
		}
		return s.writeResult(request.ID, toolResult("ok", false))
	default:
		return s.writeError(request.ID, -32601, "method not found")
	}
}

func (s Server) read(call toolCall) (map[string]any, bool) {
	if call.Name != "cubase.get_selected_track" && call.Name != "cubase.get_eq" && call.Name != "cubase.list_inserts" && call.Name != "cubase.get_insert_parameters" {
		return nil, false
	}
	if s.State == nil || !s.State.Snapshot().ReadbackAvailable {
		return toolResult("no Cubase feedback received yet; reload the MIDI Remote script with the desired track selected", true), true
	}
	snapshot := s.State.Snapshot()
	switch call.Name {
	case "cubase.get_selected_track":
		return toolJSON(map[string]any{"mixer": snapshot.Mixer, "lastTransportEvent": snapshot.LastTransportEvent, "lastFeedbackAt": snapshot.LastFeedbackAt}), true
	case "cubase.get_eq":
		return toolJSON(snapshot.EQ), true
	case "cubase.list_inserts":
		return toolJSON(snapshot.Inserts), true
	case "cubase.get_insert_parameters":
		slot, err := number(call.Arguments, "slot", 1, 8)
		if err != nil || math.Trunc(slot) != slot {
			return toolResult("slot must be an integer from 1 to 8", true), true
		}
		for _, insert := range snapshot.Inserts {
			if insert.Slot == int(slot) {
				return toolJSON(insert.Parameters), true
			}
		}
		return toolResult("no feedback for requested insert slot", true), true
	}
	return nil, false
}

func (s Server) call(call toolCall) error {
	if s.Controller == nil {
		return fmt.Errorf("Cubase controller is unavailable")
	}
	switch call.Name {
	case "cubase.play":
		return s.Controller.Play()
	case "cubase.stop":
		return s.Controller.Stop()
	case "cubase.record":
		return s.Controller.Record()
	case "cubase.set_volume":
		value, err := number(call.Arguments, "value", 0, 1)
		if err != nil {
			return err
		}
		return s.Controller.SetSelectedTrackVolume(value)
	case "cubase.set_pan":
		value, err := number(call.Arguments, "value", -1, 1)
		if err != nil {
			return err
		}
		return s.Controller.SetSelectedTrackPan(value)
	case "cubase.set_mute":
		enabled, err := boolean(call.Arguments, "enabled")
		if err != nil {
			return err
		}
		return s.Controller.SetSelectedTrackMute(enabled)
	case "cubase.set_solo":
		enabled, err := boolean(call.Arguments, "enabled")
		if err != nil {
			return err
		}
		return s.Controller.SetSelectedTrackSolo(enabled)
	case "cubase.set_eq_band":
		bandValue, err := number(call.Arguments, "band", 1, 4)
		if err != nil || math.Trunc(bandValue) != bandValue {
			return fmt.Errorf("band must be an integer from 1 to 4")
		}
		parameter, err := text(call.Arguments, "parameter")
		if err != nil {
			return err
		}
		value, err := number(call.Arguments, "value", -24, 20000)
		if err != nil {
			return err
		}
		band := int(bandValue)
		switch parameter {
		case "gain_db":
			return s.Controller.SetEQGain(band, value)
		case "frequency_hz":
			return s.Controller.SetEQFrequency(band, value)
		case "q":
			return s.Controller.SetEQQ(band, value)
		case "enabled":
			return s.Controller.EnableEQBand(band, value >= 0.5)
		default:
			return fmt.Errorf("parameter must be gain_db, frequency_hz, q, or enabled")
		}
	case "cubase.bypass_insert":
		slot, err := number(call.Arguments, "slot", 1, 8)
		if err != nil || math.Trunc(slot) != slot {
			return fmt.Errorf("slot must be an integer from 1 to 8")
		}
		bypass, err := boolean(call.Arguments, "bypass")
		if err != nil {
			return err
		}
		return s.Controller.SetInsertBypass(int(slot), bypass)
	case "cubase.set_insert_parameter":
		slot, err := number(call.Arguments, "slot", 1, 8)
		if err != nil || math.Trunc(slot) != slot {
			return fmt.Errorf("slot must be an integer from 1 to 8")
		}
		parameterID, err := text(call.Arguments, "parameter_id")
		if err != nil {
			return err
		}
		value, err := number(call.Arguments, "value", 0, 1)
		if err != nil {
			return err
		}
		return s.Controller.SetInsertParameter(int(slot), parameterID, value)
	default:
		return fmt.Errorf("unknown tool: %s", call.Name)
	}
}

func number(arguments map[string]json.RawMessage, name string, min, max float64) (float64, error) {
	var value float64
	if raw, ok := arguments[name]; !ok || json.Unmarshal(raw, &value) != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < min || value > max {
		return 0, fmt.Errorf("%s must be a number from %g to %g", name, min, max)
	}
	return value, nil
}

func boolean(arguments map[string]json.RawMessage, name string) (bool, error) {
	var value bool
	if raw, ok := arguments[name]; !ok || json.Unmarshal(raw, &value) != nil {
		return false, fmt.Errorf("%s must be a boolean", name)
	}
	return value, nil
}

func text(arguments map[string]json.RawMessage, name string) (string, error) {
	var value string
	if raw, ok := arguments[name]; !ok || json.Unmarshal(raw, &value) != nil || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s must be a non-empty string", name)
	}
	return value, nil
}

func (s Server) writeResult(id json.RawMessage, result any) error {
	return s.write(map[string]any{"jsonrpc": "2.0", "id": json.RawMessage(id), "result": result})
}

func (s Server) writeError(id json.RawMessage, code int, message string) error {
	if len(id) == 0 {
		id = json.RawMessage("null")
	}
	return s.write(map[string]any{"jsonrpc": "2.0", "id": json.RawMessage(id), "error": map[string]any{"code": code, "message": message}})
}

func (s Server) write(value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(s.Out, string(encoded))
	return err
}

func toolResult(message string, isError bool) map[string]any {
	return map[string]any{"content": []map[string]string{{"type": "text", "text": message}}, "isError": isError}
}

func toolJSON(value any) map[string]any {
	encoded, err := json.Marshal(value)
	if err != nil {
		return toolResult(err.Error(), true)
	}
	return map[string]any{"content": []map[string]string{{"type": "text", "text": string(encoded)}}, "structuredContent": value, "isError": false}
}

func tools() []map[string]any {
	return []map[string]any{
		tool("cubase.play", "Start Cubase transport.", map[string]any{}),
		tool("cubase.stop", "Stop Cubase transport.", map[string]any{}),
		tool("cubase.record", "Toggle Cubase transport recording. Use only in a disposable project.", map[string]any{}),
		tool("cubase.get_selected_track", "Get event-derived selected-track mixer state. The track name is not available through the current MIDI Remote bridge.", map[string]any{}),
		tool("cubase.get_eq", "Get event-derived values for the selected track's four EQ bands.", map[string]any{}),
		tool("cubase.list_inserts", "List event-derived inserts on the selected track.", map[string]any{}),
		tool("cubase.get_insert_parameters", "Get event-derived parameter data for an insert slot.", map[string]any{"slot": numberSchema(1, 8)}),
		tool("cubase.set_volume", "Set the selected track volume as a normalized value.", map[string]any{"value": numberSchema(0, 1)}),
		tool("cubase.set_pan", "Set the selected track pan from -1 (left) to 1 (right).", map[string]any{"value": numberSchema(-1, 1)}),
		tool("cubase.set_mute", "Set mute on the selected track.", map[string]any{"enabled": map[string]any{"type": "boolean"}}),
		tool("cubase.set_solo", "Set solo on the selected track.", map[string]any{"enabled": map[string]any{"type": "boolean"}}),
		tool("cubase.set_eq_band", "Set one precise selected-track EQ band parameter.", map[string]any{"band": numberSchema(1, 4), "parameter": map[string]any{"type": "string", "enum": []string{"gain_db", "frequency_hz", "q", "enabled"}}, "value": numberSchema(-24, 20000)}),
		tool("cubase.bypass_insert", "Set bypass for an insert slot on the selected track.", map[string]any{"slot": numberSchema(1, 8), "bypass": map[string]any{"type": "boolean"}}),
		tool("cubase.set_insert_parameter", "Set normalized generic parameter 1–8 for insert slot 1.", map[string]any{"slot": numberSchema(1, 8), "parameter_id": map[string]any{"type": "string"}, "value": numberSchema(0, 1)}),
	}
}

func tool(name, description string, properties map[string]any) map[string]any {
	required := make([]string, 0, len(properties))
	for name := range properties {
		required = append(required, name)
	}
	return map[string]any{"name": name, "description": description, "inputSchema": map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}}
}

func numberSchema(min, max float64) map[string]any {
	return map[string]any{"type": "number", "minimum": min, "maximum": max}
}
