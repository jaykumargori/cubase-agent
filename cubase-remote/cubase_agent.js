// Cubase MIDI Remote API v1.0+; documented transport proof for Cubase 15.
var midiremote_api = require('midiremote_api_v1')
var deviceDriver = midiremote_api.makeDeviceDriver('LocalFirst', 'Cubase Agent', 'LocalFirst')
var midiInput = deviceDriver.mPorts.makeMidiInput('agentInput')
var midiOutput = deviceDriver.mPorts.makeMidiOutput('agentOutput')
deviceDriver.makeDetectionUnit().detectPortPair(midiInput, midiOutput)
  .expectInputNameContains('CubaseAgent Out')
  .expectOutputNameContains('CubaseAgent In')
var play = deviceDriver.mSurface.makeButton(0, 0, 2, 1)
var stop = deviceDriver.mSurface.makeButton(2, 0, 2, 1)
var record = deviceDriver.mSurface.makeButton(4, 0, 2, 1)
play.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 115)
stop.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 114)
record.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 113)
// Mapping pages are exclusive in Cubase: only the page currently selected in
// MIDI Remote receives input. Keep every non-overlapping control on one page
// so the command-line/MCP bridge remains active even while its UI is showing
// the plugin section.
var controlPage = deviceDriver.mMapping.makePage('Cubase Agent')
controlPage.makeCommandBinding(play.mSurfaceValue, 'Transport', 'Start')
controlPage.makeCommandBinding(stop.mSurfaceValue, 'Transport', 'Stop')
controlPage.makeCommandBinding(record.mSurfaceValue, 'Transport', 'Record')
var volume = deviceDriver.mSurface.makeFader(0, 2, 2, 6).setTypeVertical()
var pan = deviceDriver.mSurface.makeKnob(3, 2, 2, 2)
var mute = deviceDriver.mSurface.makeButton(5, 2, 2, 1)
var solo = deviceDriver.mSurface.makeButton(7, 2, 2, 1)
var selected = controlPage.mHostAccess.mTrackSelection.mMixerChannel
volume.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 20).setTypeAbsolute()
pan.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 21).setTypeAbsolute()
mute.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 22).setTypeAbsolute()
solo.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 23).setTypeAbsolute()
controlPage.makeValueBinding(volume.mSurfaceValue, selected.mValue.mVolume)
controlPage.makeValueBinding(pan.mSurfaceValue, selected.mValue.mPan)
controlPage.makeValueBinding(mute.mSurfaceValue, selected.mValue.mMute)
controlPage.makeValueBinding(solo.mSurfaceValue, selected.mValue.mSolo)
var eq = selected.mChannelEQ
for (var bi = 0; bi < 4; bi++) {
  var band = eq['mBand' + (bi + 1)]
  var gain = deviceDriver.mSurface.makeKnob(0 + bi * 2, 10, 2, 2)
  var freq = deviceDriver.mSurface.makeKnob(0 + bi * 2, 13, 2, 2)
  var quality = deviceDriver.mSurface.makeKnob(0 + bi * 2, 16, 2, 2)
  var on = deviceDriver.mSurface.makeButton(0 + bi * 2, 19, 2, 1)
  gain.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 40 + bi * 4).setTypeAbsolute()
  freq.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 41 + bi * 4).setTypeAbsolute()
  quality.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 42 + bi * 4).setTypeAbsolute()
  on.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 43 + bi * 4).setTypeAbsolute()
  controlPage.makeValueBinding(gain.mSurfaceValue, band.mGain)
  controlPage.makeValueBinding(freq.mSurfaceValue, band.mFreq)
  controlPage.makeValueBinding(quality.mSurfaceValue, band.mQ)
  controlPage.makeValueBinding(on.mSurfaceValue, band.mOn)
}
var insertViewer = selected.mInsertAndStripEffects.makeInsertEffectViewer('Selected Track Inserts')
var insertIdentities = []
var parameterIdentities = []
function sendTextEvent(activeDevice, eventType, slot, parameter, parts) {
  var message = [0xf0, 0x7d, 0x43, 0x41, 1, eventType, slot, parameter]
  for (var partIndex = 0; partIndex < parts.length; partIndex++) {
    var text = String(parts[partIndex] || '')
    for (var charIndex = 0; charIndex < text.length && message.length < 254; charIndex++) {
      var code = text.charCodeAt(charIndex)
      message.push(code >= 32 && code <= 126 ? code : 63)
    }
    if (partIndex + 1 < parts.length) message.push(0)
  }
  message.push(0xf7)
  midiOutput.sendMidi(activeDevice, message)
}
function bindInsertIdentity(viewer, slot) {
  viewer.mOnChangePluginIdentity = function (activeDevice, activeMapping, pluginName, pluginVendor, pluginVersion, formatVersion) {
    insertIdentities[slot - 1] = [pluginName, pluginVendor, pluginVersion, formatVersion]
    sendTextEvent(activeDevice, 1, slot, 0, insertIdentities[slot - 1])
  }
}
for (var ii = 0; ii < 8; ii++) {
  var insertSlot = insertViewer.accessSlotAtIndex(ii)
  bindInsertIdentity(insertSlot, ii + 1)
  var bypass = deviceDriver.mSurface.makeButton(ii % 4 * 2, 24 + Math.floor(ii / 4) * 3, 2, 1)
  bypass.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 60 + ii).setTypeAbsolute()
  controlPage.makeValueBinding(bypass.mSurfaceValue, insertSlot.mBypass)
}
var pluginSlot = selected.mInsertAndStripEffects.makeInsertEffectViewer('Selected Insert Parameters').accessSlotAtIndex(0)
function bindParameterIdentity(parameterValue, parameterIndex) {
  parameterValue.mOnTitleChange = function (activeDevice, activeMapping, objectTitle, valueTitle) {
    parameterIdentities[parameterIndex - 1] = valueTitle || objectTitle
    sendTextEvent(activeDevice, 2, 1, parameterIndex, [parameterIdentities[parameterIndex - 1]])
  }
  parameterValue.mOnDisplayValueChange = function (activeDevice, activeMapping, value, units) {
    sendTextEvent(activeDevice, 3, 1, parameterIndex, [String(value) + (units ? ' ' + units : '')])
  }
}
for (var pi = 0; pi < 8; pi++) {
  var parameter = pluginSlot.mParameterBankZone.makeParameterValue()
  bindParameterIdentity(parameter, pi + 1)
  var parameterKnob = deviceDriver.mSurface.makeKnob(pi * 2, 30, 2, 2)
  parameterKnob.mSurfaceValue.mMidiBinding.setInputPort(midiInput).setOutputPort(midiOutput).bindToControlChange(0, 80 + pi).setTypeAbsolute()
  controlPage.makeValueBinding(parameterKnob.mSurfaceValue, parameter)
}
