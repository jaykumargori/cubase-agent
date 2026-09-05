package main

import (
	"context"
	"fmt"
	"github.com/jaykumargori/cubase-agent/internal/cubase"
	"github.com/jaykumargori/cubase-agent/internal/mcp"
	"github.com/jaykumargori/cubase-agent/internal/midi"
	"github.com/jaykumargori/cubase-agent/internal/protocol"
	"github.com/jaykumargori/cubase-agent/internal/state"
	"math"
	"os"
	"strconv"
	"time"
)

const port = "CubaseAgent Out"
const feedbackPort = "CubaseAgent In"

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	c, e := openMIDI(port, 3*time.Second)
	if e != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n\nExpected: %s\nSee: docs/macos-midi-setup.md\n", e, port)
		os.Exit(1)
	}
	defer c.Close()
	b := cubase.Bridge{MIDI: c}
	switch os.Args[1] {
	case "mcp":
		store := state.New()
		go state.Pump(context.Background(), c, store)
		if err := (mcp.Server{Controller: &b, State: store, In: os.Stdin, Out: os.Stdout}).Run(); err != nil {
			fmt.Fprintln(os.Stderr, "[ERROR] MCP server:", err)
			os.Exit(1)
		}
		return
	case "play":
		report(b.Play(), "transport.play")
	case "stop":
		report(b.Stop(), "transport.stop")
	case "record":
		report(b.Record(), "transport.record")
	case "status":
		fmt.Println("[INFO] MIDI connected:", port)
		fmt.Println("[INFO] MIDI feedback connected:", feedbackPort)
		return
	case "feedback":
		frame, err := c.Receive(2 * time.Second)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[ERROR]", err)
			os.Exit(1)
		}
		feedback, err := protocol.DecodeFeedback(frame)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[ERROR]", err)
			os.Exit(1)
		}
		fmt.Printf("[INFO] feedback type=%s value=%.4f band=%d slot=%d parameter=%d name=%q vendor=%q display=%q\n", feedback.Type, feedback.Normalized, feedback.Band, feedback.Slot, feedback.Parameter, feedback.Name, feedback.Vendor, feedback.DisplayValue)
		return
	case "volume":
		report(b.SetSelectedTrackVolume(parseNumber(os.Args, 2, 0, 1)), "track.volume.set")
	case "pan":
		report(b.SetSelectedTrackPan(parseNumber(os.Args, 2, -1, 1)), "track.pan.set")
	case "mute":
		report(b.SetSelectedTrackMute(onOff(os.Args, 2)), "track.mute.set")
	case "solo":
		report(b.SetSelectedTrackSolo(onOff(os.Args, 2)), "track.solo.set")
	case "insert-bypass":
		if len(os.Args) < 4 {
			usage()
			os.Exit(2)
		}
		slot, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "slot must be 1..8")
			os.Exit(2)
		}
		report(b.SetInsertBypass(slot, onOff(os.Args, 3)), "insert.bypass.set")
	case "insert":
		handleInsert(os.Args, &b, c)
		return
	case "inserts":
		reportInsertSnapshot(c, 0)
		return
	case "plugin-param":
		if len(os.Args) < 4 {
			usage()
			os.Exit(2)
		}
		parameter, err := strconv.Atoi(os.Args[2])
		if err != nil || parameter < 1 || parameter > 8 {
			fmt.Fprintln(os.Stderr, "plugin parameter must be 1..8")
			os.Exit(2)
		}
		report(b.SetInsertParameter(1, strconv.Itoa(parameter), parseNumber(os.Args, 3, 0, 1)), "insert.parameter.set")
	case "eq":
		if len(os.Args) < 5 {
			usage()
			os.Exit(2)
		}
		band, err := strconv.Atoi(os.Args[3])
		if err != nil || band < 1 || band > 4 {
			fmt.Fprintln(os.Stderr, "band must be 1..4")
			os.Exit(2)
		}
		switch os.Args[2] {
		case "gain":
			v := parseEQValue(os.Args[4])
			report(b.SetEQGain(band, v), "eq.band.gain.set")
		case "freq", "frequency":
			v := parseEQValue(os.Args[4])
			report(b.SetEQFrequency(band, v), "eq.band.frequency.set")
		case "q":
			v := parseEQValue(os.Args[4])
			report(b.SetEQQ(band, v), "eq.band.q.set")
		case "enable":
			report(b.EnableEQBand(band, onOff(os.Args, 4)), "eq.band.enable")
		default:
			usage()
			os.Exit(2)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func parseEQValue(value string) float64 {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "value must be numeric")
		os.Exit(2)
	}
	return v
}

type midiOpenResult struct {
	client *midi.Client
	err    error
}

func openMIDI(name string, timeout time.Duration) (*midi.Client, error) {
	result := make(chan midiOpenResult, 1)
	go func() {
		client, err := midi.OpenDuplex(name, feedbackPort)
		result <- midiOpenResult{client: client, err: err}
	}()
	select {
	case opened := <-result:
		return opened.client, opened.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("CoreMIDI connection timed out after %s", timeout)
	}
}

func parseNumber(args []string, i int, min, max float64) float64 {
	if len(args) <= i {
		fmt.Fprintln(os.Stderr, "missing value")
		os.Exit(2)
	}
	v, e := strconv.ParseFloat(args[i], 64)
	if e != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < min || v > max {
		fmt.Fprintln(os.Stderr, "value out of range")
		os.Exit(2)
	}
	return v
}
func onOff(args []string, i int) bool {
	if len(args) <= i || (args[i] != "on" && args[i] != "off") {
		fmt.Fprintln(os.Stderr, "use on|off")
		os.Exit(2)
	}
	if args[i] == "on" {
		return true
	}
	return false
}

func report(err error, command string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR]", err)
		os.Exit(1)
	}
	fmt.Println("[INFO] command", command, "sent")
}
func handleInsert(args []string, b *cubase.Bridge, client *midi.Client) {
	if len(args) < 4 {
		usage()
		os.Exit(2)
	}
	slot, err := strconv.Atoi(args[2])
	if err != nil || slot < 1 || slot > 8 {
		fmt.Fprintln(os.Stderr, "insert slot must be 1..8")
		os.Exit(2)
	}
	switch args[3] {
	case "bypass":
		report(b.SetInsertBypass(slot, onOff(args, 4)), "insert.bypass.set")
	case "params":
		reportInsertSnapshot(client, slot)
		return
	case "set":
		if slot != 1 {
			fmt.Fprintln(os.Stderr, "[ERROR] generic parameter bank is currently mapped only for insert slot 1")
			os.Exit(3)
		}
		if len(args) < 6 {
			usage()
			os.Exit(2)
		}
		parameter, parseErr := strconv.Atoi(args[4])
		if parseErr != nil || parameter < 1 || parameter > 8 {
			fmt.Fprintln(os.Stderr, "plugin parameter must be 1..8")
			os.Exit(2)
		}
		report(b.SetInsertParameter(slot, strconv.Itoa(parameter), parseNumber(args, 5, 0, 1)), "insert.parameter.set")
	default:
		usage()
		os.Exit(2)
	}
}

func reportInsertSnapshot(client *midi.Client, requestedSlot int) {
	deadline := time.Now().Add(3 * time.Second)
	events := make([]protocol.Feedback, 0, 16)
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining > 250*time.Millisecond {
			remaining = 250 * time.Millisecond
		}
		frame, err := client.Receive(remaining)
		if err != nil {
			continue
		}
		feedback, err := protocol.DecodeFeedback(frame)
		if err == nil {
			events = append(events, feedback)
		}
	}
	snapshots := cubase.BuildInsertSnapshots(events)
	found := false
	for _, snapshot := range snapshots {
		if requestedSlot != 0 && snapshot.Slot != requestedSlot {
			continue
		}
		found = true
		if requestedSlot == 0 {
			fmt.Printf("[INFO] insert slot=%d name=%q vendor=%q version=%q format=%q bypassed=%t\n", snapshot.Slot, snapshot.Name, snapshot.Vendor, snapshot.Version, snapshot.Format, snapshot.Bypassed)
			continue
		}
		fmt.Printf("[INFO] insert slot=%d name=%q vendor=%q bypassed=%t\n", snapshot.Slot, snapshot.Name, snapshot.Vendor, snapshot.Bypassed)
		for _, parameter := range snapshot.Parameters {
			fmt.Printf("[INFO] parameter id=%s name=%q value=%.4f display=%q\n", parameter.ID, parameter.Name, parameter.NormalizedValue, parameter.DisplayValue)
		}
	}
	if !found {
		fmt.Fprintln(os.Stderr, "[ERROR] no insert feedback received; reload the Cubase MIDI Remote script with the target track selected, then retry")
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("cubase-agent mcp|status|feedback|play|stop|record|volume 0..1|pan -1..1|mute on|off|solo on|off|eq <gain|freq|q|enable> <band> <value>|inserts|insert <slot> params|bypass on|off|insert 1 set <parameter 1..8> <value 0..1>")
}
