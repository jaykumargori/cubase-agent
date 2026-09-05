# Cubase API notes

Cubase 15 MIDI Remote API mappings currently used:

- `mTrackSelection.mMixerChannel.mValue.mVolume`, `mPan`, `mMute`, `mSolo`
- `mTrackSelection.mMixerChannel.mChannelEQ.mBand1..mBand4`
- `mInsertAndStripEffects.makeInsertEffectViewer(...)`
- insert slot access through `accessSlotAtIndex(index).mBypass`
- generic parameter-bank access through `mParameterBankZone.makeParameterValue()`
- MIDI feedback through `SurfaceValueMidiBinding.setOutputPort(...)`
- `mOnChangePluginIdentity`, `mOnTitleChange`, and `mOnDisplayValueChange`
  publish insert identity and generic parameter metadata to Go as a compact
  manufacturer-specific SysEx event.

Transport, mixer, EQ, insert bypass, and generic parameter-bank writes were
smoke-tested through CoreMIDI. Insert identity and generic parameter discovery
is event-driven: reload the script with the desired track selected, then invoke
the CLI discovery command within three seconds. It does not rely on unverified
DirectAccess methods.
