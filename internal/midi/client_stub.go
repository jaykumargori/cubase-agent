//go:build !darwin

package midi

import "errors"
import "time"

type Client struct{}

func Open(string) (*Client, error) { return nil, errors.New("CoreMIDI is only supported on macOS") }
func OpenDuplex(string, string) (*Client, error) {
	return nil, errors.New("CoreMIDI is only supported on macOS")
}
func (c *Client) Send([]byte) error { return errors.New("CoreMIDI is only supported on macOS") }
func (c *Client) Receive(time.Duration) ([]byte, error) {
	return nil, errors.New("CoreMIDI is only supported on macOS")
}
