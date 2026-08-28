# macOS MIDI setup

1. Open **Audio MIDI Setup** → Window → Show MIDI Studio.
2. Double-click **IAC Driver**, enable **Device is online**, and add two buses:
   `CubaseAgent Out` and `CubaseAgent In`.
3. In Cubase, open Studio Setup → MIDI Remote Manager and add the supplied
   `cubase-remote/cubase_agent.js` after its API declarations have been
   verified against the installed Cubase 15 Programmer's Guide.

The CLI sends commands to `CubaseAgent Out` and listens for Cubase feedback on
`CubaseAgent In`. Both ports are required by `cubase-agent status`.
