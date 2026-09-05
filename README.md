# Cubase Agent

Local-first Go/CoreMIDI bridge for Cubase Pro 15. It sends transport, selected
track mixer, four-band channel EQ, insert bypass, and a generic eight-parameter
plugin bank to Cubase. Commands use `CubaseAgent Out`; feedback uses
`CubaseAgent In`.

Build with `go build ./cmd/cubase-agent`.
Run `cubase-agent play` or `cubase-agent stop` after enabling the documented
IAC buses. Mixer, EQ, insert bypass, plugin-bank writes, MIDI feedback input,
typed feedback decoding, request matching, and generic plugin identity and
parameter-name feedback, and a local stdio MCP server are implemented. AI is
not implemented. EQ
frequency and Q use logarithmic 7-bit MIDI normalization.

Plugin parameter test: `cubase-agent plugin-param 1 0.5` (parameters 1–8,
normalized 0–1).

## MCP

Run `cubase-agent mcp` to expose the reversible transport, selected-track
mixer, EQ, and insert-write controls over a local stdio MCP server. It uses no
network service and does not save, close, delete, render, or export projects.

## License

MIT
