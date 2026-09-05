//go:build darwin

package midi

/*
#cgo LDFLAGS: -framework CoreMIDI -framework CoreFoundation
#include <CoreMIDI/CoreMIDI.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
int ca_open(const char*, MIDIClientRef*, MIDIPortRef*, MIDIEndpointRef*);
int ca_connect_input(MIDIClientRef, const char*, MIDIPortRef*, MIDIEndpointRef*);
int ca_send(MIDIPortRef, MIDIEndpointRef, const unsigned char*, int);
int ca_receive(unsigned char*, int);
void ca_close(MIDIClientRef, MIDIPortRef, MIDIPortRef);
*/
import "C"
import (
	"errors"
	"time"
	"unsafe"
)

type Client struct {
	client    C.MIDIClientRef
	port      C.MIDIPortRef
	endpoint  C.MIDIEndpointRef
	inputPort C.MIDIPortRef
	source    C.MIDIEndpointRef
}

func OpenDuplex(outputName, inputName string) (*Client, error) {
	client, err := Open(outputName)
	if err != nil {
		return nil, err
	}
	name := C.CString(inputName)
	defer C.free(unsafe.Pointer(name))
	if C.ca_connect_input(client.client, name, &client.inputPort, &client.source) != 0 {
		return nil, errors.New("MIDI feedback port not found: " + inputName)
	}
	return client, nil
}

func Open(name string) (*Client, error) {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	var c C.MIDIClientRef
	var p C.MIDIPortRef
	var e C.MIDIEndpointRef
	if C.ca_open(n, &c, &p, &e) != 0 {
		return nil, errors.New("MIDI port not found: " + name)
	}
	return &Client{client: c, port: p, endpoint: e}, nil
}
func (c *Client) Receive(timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)
	for {
		buffer := make([]byte, 256)
		length := int(C.ca_receive((*C.uchar)(unsafe.Pointer(&buffer[0])), C.int(len(buffer))))
		if length > 0 {
			return buffer[:length], nil
		}
		if time.Now().After(deadline) {
			return nil, errors.New("MIDI feedback timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
func (c *Client) Send(b []byte) error {
	if C.ca_send(c.port, c.endpoint, (*C.uchar)(unsafe.Pointer(&b[0])), C.int(len(b))) != 0 {
		return errors.New("CoreMIDI send failed")
	}
	return nil
}

func (c *Client) Close() error {
	if c == nil || c.client == 0 {
		return nil
	}
	C.ca_close(c.client, c.port, c.inputPort)
	c.client = 0
	c.port = 0
	c.inputPort = 0
	return nil
}
