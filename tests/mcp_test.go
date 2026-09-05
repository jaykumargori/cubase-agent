package tests

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jaykumargori/cubase-agent/internal/mcp"
)

type fakeController struct {
	called string
	value  float64
}

func (f *fakeController) Play() error   { f.called = "play"; return nil }
func (f *fakeController) Stop() error   { f.called = "stop"; return nil }
func (f *fakeController) Record() error { f.called = "record"; return nil }
func (f *fakeController) SetSelectedTrackVolume(v float64) error {
	f.called, f.value = "volume", v
	return nil
}
func (f *fakeController) SetSelectedTrackPan(v float64) error {
	f.called, f.value = "pan", v
	return nil
}
func (f *fakeController) SetSelectedTrackMute(bool) error   { f.called = "mute"; return nil }
func (f *fakeController) SetSelectedTrackSolo(bool) error   { f.called = "solo"; return nil }
func (f *fakeController) SetEQGain(int, float64) error      { f.called = "eq gain"; return nil }
func (f *fakeController) SetEQFrequency(int, float64) error { f.called = "eq frequency"; return nil }
func (f *fakeController) SetEQQ(int, float64) error         { f.called = "eq q"; return nil }
func (f *fakeController) EnableEQBand(int, bool) error      { f.called = "eq enabled"; return nil }
func (f *fakeController) SetInsertBypass(int, bool) error   { f.called = "bypass"; return nil }
func (f *fakeController) SetInsertParameter(int, string, float64) error {
	f.called = "parameter"
	return nil
}

func TestMCPInitializeListAndCall(t *testing.T) {
	input := strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{}}\n" +
		"{\"jsonrpc\":\"2.0\",\"method\":\"notifications/initialized\"}\n" +
		"{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/list\",\"params\":{}}\n" +
		"{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"cubase.set_volume\",\"arguments\":{\"value\":0.7}}}\n")
	output := &bytes.Buffer{}
	controller := &fakeController{}
	if err := (mcp.Server{Controller: controller, In: input, Out: output}).Run(); err != nil {
		t.Fatal(err)
	}
	if controller.called != "volume" || controller.value != 0.7 {
		t.Fatalf("unexpected controller call: %+v", controller)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("responses = %d: %s", len(lines), output.String())
	}
	var listed struct {
		Result struct {
			Tools []json.RawMessage `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &listed); err != nil || len(listed.Result.Tools) < 10 {
		t.Fatalf("unexpected tools/list response: %s (%v)", lines[1], err)
	}
	if !strings.Contains(lines[2], "\"isError\":false") {
		t.Fatalf("unexpected tool result: %s", lines[2])
	}
}

func TestMCPRejectsInvalidToolArguments(t *testing.T) {
	input := strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"cubase.set_volume\",\"arguments\":{\"value\":2}}}\n")
	output := &bytes.Buffer{}
	if err := (mcp.Server{Controller: &fakeController{}, In: input, Out: output}).Run(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "\"isError\":true") {
		t.Fatalf("invalid argument did not return a tool error: %s", output.String())
	}
}
