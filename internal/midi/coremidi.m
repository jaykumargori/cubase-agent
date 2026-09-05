#include <CoreMIDI/CoreMIDI.h>
#include <CoreFoundation/CoreFoundation.h>
#include <string.h>
#include <stdlib.h>
#include <pthread.h>
typedef struct { unsigned int len; unsigned char data[256]; } CAMessage;
static CAMessage ca_queue[64];
static unsigned int ca_head = 0, ca_tail = 0;
static pthread_mutex_t ca_lock = PTHREAD_MUTEX_INITIALIZER;
// Audio MIDI Setup exposes IAC buses as "IAC Driver <bus name>" on some
// macOS versions, while other CoreMIDI clients report only the bus name.
// Prefer an exact match, then accept the configured bus name as a suffix.
static int ca_match(CFStringRef s, const char *want) {
  char buf[256];
  if (!s || !CFStringGetCString(s, buf, sizeof(buf), kCFStringEncodingUTF8)) return 0;
  size_t actualLen = strlen(buf), wantedLen = strlen(want);
  return strcmp(buf, want) == 0 || (actualLen > wantedLen && strcmp(buf + actualLen - wantedLen, want) == 0);
}
static void ca_read(const MIDIPacketList *list, void *context, void *sourceContext) {
  const MIDIPacket *packet = &list->packet[0];
  for (UInt32 i = 0; i < list->numPackets; i++) {
    if (packet->length > 0) {
      pthread_mutex_lock(&ca_lock);
      unsigned int next = (ca_head + 1) % 64;
      if (next != ca_tail) {
        unsigned int len = packet->length > 256 ? 256 : packet->length;
        ca_queue[ca_head].len = len;
        memcpy(ca_queue[ca_head].data, packet->data, len);
        ca_head = next;
      }
      pthread_mutex_unlock(&ca_lock);
    }
    packet = MIDIPacketNext(packet);
  }
}
int ca_open(const char *name, MIDIClientRef *client, MIDIPortRef *port, MIDIEndpointRef *endpoint) {
  if (MIDIClientCreate(CFSTR("Cubase Agent"), NULL, NULL, client) != 0 || MIDIOutputPortCreate(*client, CFSTR("Output"), port) != 0) return -1;
  ItemCount n = MIDIGetNumberOfDestinations();
  for (ItemCount i = 0; i < n; i++) { MIDIEndpointRef d = MIDIGetDestination(i); CFStringRef s = NULL; MIDIObjectGetStringProperty(d, kMIDIPropertyName, &s); int ok = ca_match(s, name); if (s) CFRelease(s); if (ok) { *endpoint = d; return 0; } }
  return -2;
}
int ca_connect_input(MIDIClientRef client, const char *name, MIDIPortRef *port, MIDIEndpointRef *endpoint) {
  if (MIDIInputPortCreate(client, CFSTR("Input"), ca_read, NULL, port) != 0) return -1;
  ItemCount n = MIDIGetNumberOfSources();
  for (ItemCount i = 0; i < n; i++) { MIDIEndpointRef sref = MIDIGetSource(i); CFStringRef s = NULL; MIDIObjectGetStringProperty(sref, kMIDIPropertyName, &s); int ok = ca_match(s, name); if (s) CFRelease(s); if (ok) { *endpoint = sref; return MIDIPortConnectSource(*port, sref, NULL); } }
  return -2;
}
int ca_receive(unsigned char *bytes, int maxLen) {
  pthread_mutex_lock(&ca_lock);
  if (ca_head == ca_tail) { pthread_mutex_unlock(&ca_lock); return 0; }
  unsigned int len = ca_queue[ca_tail].len;
  if (len > (unsigned int)maxLen) len = maxLen;
  memcpy(bytes, ca_queue[ca_tail].data, len); ca_tail = (ca_tail + 1) % 64;
  pthread_mutex_unlock(&ca_lock);
  return len;
}
int ca_send(MIDIPortRef port, MIDIEndpointRef endpoint, const unsigned char *bytes, int len) {
  Byte buffer[256]; if (len > (int)sizeof(buffer) - 100) return -1; MIDIPacketList *list = (MIDIPacketList *)buffer; MIDIPacket *packet = MIDIPacketListInit(list); packet = MIDIPacketListAdd(list, sizeof(buffer), packet, 0, len, bytes); return packet ? MIDISend(port, endpoint, list) : -2;
}
void ca_close(MIDIClientRef client, MIDIPortRef outputPort, MIDIPortRef inputPort) {
  if (inputPort) MIDIPortDispose(inputPort);
  if (outputPort) MIDIPortDispose(outputPort);
  if (client) MIDIClientDispose(client);
}
