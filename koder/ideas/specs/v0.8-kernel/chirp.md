# Chirp (Audio Data Transfer)

## Summary

Encode data as audio, decode audio back to data. Physical-layer communication
when no network exists at all. Last-resort data transfer via speaker and
microphone - like a modem for the apocalypse.

## Why Kernel-Level

Chirp is the escape hatch when all else fails:
- No internet, no LAN, no Bluetooth, no cables
- Two devices within earshot can still exchange data
- Share encryption keys across an air gap
- Bootstrap trust when you can't network

Unlike Beacon and Timekeeper, Chirp is **explicit only**. It doesn't
integrate into other subsystems - it's a manual tool for extreme scenarios.

## The Resilience Contract

```
Network available:
  Use network → Chirp dormant

No network, physical proximity:
  User explicitly uses Chirp → audio transfer

No network, no proximity:
  Nothing works → that's physics
```

**Apps don't auto-use Chirp.** It's a conscious "break glass" action.

## Use Cases

1. **Key Exchange**: Share a public key or mesh identity with someone
   who can hear you but can't network with you

2. **Air-Gap Bridge**: Transfer small secrets into/out of air-gapped systems

3. **Emergency Sync**: Sync critical state when all networking is down

4. **QR Backup**: Audio fallback when cameras don't work

## Protocol

FSK (Frequency-Shift Keying) encoding:
- Two frequencies represent 0 and 1
- Preamble for synchronization
- Reed-Solomon error correction
- Checksum for integrity

```
Audio Frame:
┌──────────┬──────────┬──────────────────┬──────────┐
│ Preamble │ Length   │ Payload (+ ECC)  │ Checksum │
│ (sync)   │ (2 bytes)│ (variable)       │ (4 bytes)│
└──────────┴──────────┴──────────────────┴──────────┘
```

## Bandwidth

Intentionally slow for reliability:

| Mode       | Speed          | Best For                    |
| ---------- | -------------- | --------------------------- |
| `robust`   | ~50 bytes/sec  | Noisy environments          |
| `standard` | ~200 bytes/sec | Normal conditions (default) |
| `fast`     | ~500 bytes/sec | Quiet, close range          |

A 256-byte key takes ~1-5 seconds depending on mode.

## CLI

```bash
# Send data via speaker
fazt chirp send "hello world"
fazt chirp send --file identity.json
fazt chirp send --file secret.key --mode robust

# Listen via microphone
fazt chirp listen
fazt chirp listen --timeout 30s --output received.json

# Encode to WAV file (no speaker needed)
fazt chirp encode "hello world" > message.wav
fazt chirp encode --file identity.json --output identity.wav

# Decode from WAV file (no mic needed)
fazt chirp decode message.wav
fazt chirp decode --file recording.wav --output data.json
```

## JS API (Explicit Only)

```javascript
// Encode data to audio buffer
const audioBuffer = await fazt.chirp.encode(data, options);
// options: { mode: 'standard' }

// Decode audio buffer to data
const data = await fazt.chirp.decode(audioBuffer);
// Throws if decode fails (noise, corruption)

// Play through speaker (requires audio permission)
await fazt.chirp.send(data, options);
// options: { mode: 'standard' }

// Listen via microphone (requires audio permission)
const data = await fazt.chirp.listen(options);
// options: { timeout: 30000, mode: 'standard' }
// Returns: Promise<Buffer> or throws on timeout/failure
```

## Typical Flow

### Sender (Your Pi)

```bash
# Export your mesh identity
fazt mesh export-identity > /tmp/identity.json

# Send via audio
fazt chirp send --file /tmp/identity.json
# [Speaker plays chirping sounds for ~3 seconds]
```

### Receiver (Their Device)

```bash
# Listen for incoming data
fazt chirp listen --timeout 30s --output /tmp/received.json
# [Microphone listens, decodes chirps]
# Received 256 bytes, saved to /tmp/received.json

# Import the identity
fazt mesh import-identity /tmp/received.json
```

## Error Handling

```bash
fazt chirp listen --timeout 10s
# Error: Decode failed - too much noise
# Error: Timeout - no valid chirp detected
# Error: Checksum mismatch - data corrupted

# Retry with robust mode
fazt chirp listen --timeout 30s --mode robust
```

## Audio Characteristics

| Parameter        | Value                                |
| ---------------- | ------------------------------------ |
| Sample rate      | 44100 Hz                             |
| Frequencies      | 1200 Hz (0), 2400 Hz (1)             |
| Bit duration     | 5-20ms (mode dependent)              |
| Preamble         | 500ms sync tone                      |
| Error correction | Reed-Solomon (can recover ~10% loss) |

Designed to survive:
- Laptop/phone speakers (not HiFi quality)
- Background room noise
- Slight echo/reverb
- Distance of 1-3 meters

## Implementation Notes

- Pure Go audio encoding/decoding
- ~400 lines of code
- No external dependencies (no cgo audio libs)
- For playback/capture: uses OS audio APIs via small shim
- WAV encode/decode is fully portable

### Headless Operation

On a Pi without speakers/mic:

```bash
# Generate WAV, transfer to phone, play there
fazt chirp encode --file data.json --output message.wav
# Copy message.wav to phone, play it near target device

# Or: record on phone, transfer WAV, decode on Pi
fazt chirp decode --file recording.wav
```

## Limits

| Limit            | Default   |
| ---------------- | --------- |
| `maxPayloadSize` | 4 KB      |
| `defaultTimeout` | 30s       |
| `maxTimeout`     | 5 minutes |
| `minFrequency`   | 1000 Hz   |
| `maxFrequency`   | 4000 Hz   |

## Security Considerations

- **Chirp is not encrypted by default** - it's transport, not security
- Encrypt sensitive data before chirping:
  ```bash
  fazt security encrypt --file secret.json | fazt chirp send
  ```
- Anyone in earshot can record and decode
- Use for key exchange, not bulk data transfer
- Consider "key + verify out-of-band" pattern

## Why Not Bluetooth/NFC/QR?

| Method    | Problem                                   |
| --------- | ----------------------------------------- |
| Bluetooth | Requires pairing, not universal           |
| NFC       | Requires hardware, very short range       |
| QR        | Camera required, size limited             |
| **Chirp** | Universal: all devices have speakers/mics |

Chirp is the lowest common denominator. It works between a Pi with
a $2 speaker and any smartphone.

## Integration with Other Primitives

Chirp intentionally doesn't auto-integrate. But it composes well:

```bash
# Share your Beacon identity
fazt beacon export | fazt chirp send

# Share a vault secret (encrypted)
fazt security vault export mykey | fazt chirp send

# Bootstrap mesh connection
fazt mesh export-identity | fazt chirp send
```

The pattern: export from one primitive → chirp send → chirp listen →
import to other device.
