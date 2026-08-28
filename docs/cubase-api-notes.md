# Cubase API notes

Cubase 15 MIDI Remote API mappings currently used:

- `mTrackSelection.mMixerChannel.mValue.mVolume`, `mPan`, `mMute`, `mSolo`
- `mTrackSelection.mMixerChannel.mChannelEQ.mBand1..mBand4`
- `mInsertAndStripEffects.makeInsertEffectViewer(...)`
- insert slot access through `accessSlotAtIndex(index).mBypass`
- generic parameter-bank access through `mParameterBankZone.makeParameterValue()`
- MIDI feedback through `SurfaceValueMidiBinding.setOutputPort(...)`
- API 1.3 DirectAccess exposes parameter introspection and
  `mPluginManager`, but runtime identity/name serialization to the Go process
  is not yet implemented

Transport, mixer, EQ, insert bypass, and generic parameter-bank writes were
smoke-tested through CoreMIDI. The dedicated feedback port and decoder are
implemented; live feedback verification requires Cubase to reload the current
two-port script. Generic plugin-name and parameter-name discovery remains
blocked on a serialization protocol for API 1.3 DirectAccess data.
