# Manual test

After installing the MIDI Remote script:

1. Enable the IAC Driver bus as described in `macos-midi-setup.md`.
2. Open Cubase Pro 15 and load the MIDI Remote script.
3. Create an empty project and confirm the bridge reports/accepts the remote.
4. Run `cubase-agent play`; transport should start.
5. Run `cubase-agent stop`; transport should stop.
6. Run `cubase-agent volume 0.5`; the selected-track fader should move.
7. Run `cubase-agent mute on`, then `cubase-agent mute off`.
8. Run `cubase-agent eq gain 1 2` and verify band 1 changes.
9. Run `cubase-agent insert 1 bypass on` and verify insert slot 1 is bypassed.
10. Run `cubase-agent feedback`, then change a mapped control in Cubase; a
    typed feedback message should be printed within two seconds.

`record` is implemented but should only be tested in a disposable project.
