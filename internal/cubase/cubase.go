package cubase

// Controller is the Cubase domain boundary. A future MCP server can depend on
// this interface without importing MIDI details.
type Controller interface {
	Play() error
	Stop() error
	Record() error
	SetSelectedTrackVolume(value float64) error
	SetSelectedTrackPan(value float64) error
	SetSelectedTrackMute(enabled bool) error
	SetSelectedTrackSolo(enabled bool) error
	SetEQGain(band int, db float64) error
	SetEQFrequency(band int, hz float64) error
	SetEQQ(band int, q float64) error
	EnableEQBand(band int, enabled bool) error
	SetInsertBypass(slot int, bypass bool) error
	SetInsertParameter(slot int, parameterID string, value float64) error
}
