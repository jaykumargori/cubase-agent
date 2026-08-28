# Cubase Agent

Local-first Go/CoreMIDI bridge for Cubase Pro 15. It sends transport, selected
track mixer, four-band channel EQ, insert bypass, and a generic eight-parameter
plugin bank to Cubase. Commands use `CubaseAgent Out`; feedback uses
`CubaseAgent In`.

Build with `go build ./cmd/cubase-agent`.
Run `cubase-agent play` or `cubase-agent stop` after enabling the documented
IAC buses. Mixer, EQ, insert bypass, plugin-bank writes, MIDI feedback input,
typed feedback decoding, request matching, and generic plugin identity and
parameter-name feedback are implemented. MCP and AI are not implemented. EQ
frequency and Q use logarithmic 7-bit MIDI normalization.

Plugin parameter test: `cubase-agent plugin-param 1 0.5` (parameters 1–8,
normalized 0–1).

## License

MIT
