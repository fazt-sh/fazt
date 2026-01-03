# Mnemonic (Human-Channel Data Exchange)

## Summary

Encode data as human-speakable word sequences, decode words back to data.
The ultimate fallback - works through any human communication channel:
phone calls, written notes, radio, memory, or messenger pigeon.

## Why Kernel-Level

Mnemonic is the escape hatch below all technology:
- No network, no local network, no physical proximity
- But you can make a phone call, send a letter, or talk on radio
- Exchange keys, connection tokens, or small secrets via voice
- Works across any channel humans can communicate through

Unlike Chirp (requires audio hardware), Mnemonic requires only human
communication - the most universal and resilient channel.

## The Resilience Stack

```
Layer 3: Network (Mesh, HTTP, Nostr)
Layer 2: Local Network (Beacon, Timekeeper)
Layer 1: Physical Digital (Chirp audio)
Layer 0: Human Channel (Mnemonic) ← this layer
```

Mnemonic is Layer 0 - when even Chirp won't work.

## Use Cases

1. **Phone Bootstrap**: Read connection token over phone call
2. **Paper Backup**: Write down identity recovery phrase
3. **Radio Exchange**: Share keys via ham radio
4. **Air-Gap Bridge**: Verbal data transfer into secure facilities
5. **Memory**: Memorize a short secret or recovery phrase

## Encoding

Uses BIP-39 wordlist (2048 English words):
- Well-known, auditable, standardized
- Phonetically distinct words (reduces transcription errors)
- Each word encodes 11 bits
- 24 words = 256 bits = 32 bytes

```
Data:        [32 bytes of binary data]
    ↓
Checksum:    Add 8-bit checksum
    ↓
Split:       264 bits → 24 × 11-bit chunks
    ↓
Lookup:      Each chunk → word from BIP-39 list
    ↓
Output:      "river castle moon bright seven table..."
```

## Word Counts

| Data Size | Words    | Use Case                        |
| --------- | -------- | ------------------------------- |
| 16 bytes  | 12 words | Short tokens, connection codes  |
| 32 bytes  | 24 words | Keys, identities, secrets       |
| 64 bytes  | 48 words | Extended data (practical limit) |

Beyond 48 words becomes unwieldy for human exchange.

## CLI

```bash
# Encode raw data
echo -n "secret data" | fazt mnemonic encode
# river castle moon bright seven table hotel alpha

# Encode file
fazt mnemonic encode --file identity.json
# (outputs 24-48 words depending on size)

# Decode words to data
fazt mnemonic decode "river castle moon bright seven table hotel alpha"
# secret data

# Decode to file
fazt mnemonic decode "river castle..." --output recovered.json

# Generate mesh invite as mnemonic
fazt mesh invite --format mnemonic
# Share this phrase: "alpha bravo charlie delta echo foxtrot..."

# Join mesh using mnemonic
fazt mesh join --mnemonic "alpha bravo charlie delta echo foxtrot..."
```

## JS API

```javascript
// Encode data to word sequence
const words = fazt.mnemonic.encode(data);
// Returns: "river castle moon bright seven table..."

// Decode words back to data
const data = fazt.mnemonic.decode(words);
// Returns: Buffer

// Validate words (check checksum, valid words)
const valid = fazt.mnemonic.validate(words);
// Returns: { valid: boolean, error?: string }

// Get word count for data size
const count = fazt.mnemonic.wordCount(byteLength);
// Returns: number of words needed
```

## Typical Flows

### Phone Call Bootstrap

```
Alice (has Fazt):
$ fazt mesh invite --format mnemonic
Share this phrase:
  "alpha bravo charlie delta echo foxtrot
   golf hotel india juliet kilo lima"

Alice reads it over the phone...

Bob (getting Fazt):
$ fazt mesh join --mnemonic "alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima"
Connected to mesh via Alice's node.
```

### Paper Backup

```
$ fazt identity export --format mnemonic
Write down and store securely:
  "river castle moon bright seven table
   hotel alpha network gamma delta sigma
   ... (24 words total)"

[Years later, new device]

$ fazt identity import --mnemonic "river castle moon..."
Identity restored.
```

### Radio Exchange

```
Operator 1 (has key):
$ fazt security vault export mykey --format mnemonic
# "oscar papa quebec romeo sierra tango..."

[Transmits via radio]

Operator 2 (receives):
$ fazt security vault import --mnemonic "oscar papa quebec romeo sierra tango..."
Key imported.
```

## Error Handling

```bash
# Invalid word
fazt mnemonic decode "river castle mxon bright"
# Error: Unknown word "mxon" at position 3. Did you mean "moon"?

# Checksum failure (transcription error)
fazt mnemonic decode "river castle moon bright seven table hotel bravo"
# Error: Checksum mismatch. Check words for transcription errors.

# Wrong word count
fazt mnemonic decode "river castle moon"
# Error: Expected 12, 24, or 48 words, got 3.
```

## Fuzzy Matching

To help with transcription errors:

```bash
fazt mnemonic decode --fuzzy "river catle moon bright..."
# Warning: "catle" not found, assuming "castle"
# Decoded successfully.
```

Fuzzy matching uses edit distance to suggest corrections.

## Integration with Other Primitives

Mnemonic composes with other export/import flows:

```bash
# Mesh identity via mnemonic
fazt mesh export-identity | fazt mnemonic encode
fazt mnemonic decode "..." | fazt mesh import-identity

# Vault secrets via mnemonic
fazt security vault export mykey | fazt mnemonic encode
fazt mnemonic decode "..." | fazt security vault import

# Beacon identity via mnemonic (for manual peer add)
fazt beacon export | fazt mnemonic encode
```

The pattern: `export | mnemonic encode` → human channel →
`mnemonic decode | import`

## Wordlist

Uses standard BIP-39 English wordlist:
- 2048 words
- All 4-8 characters
- First 4 characters unique (allows abbreviation)
- Phonetically distinct
- No offensive words

Example words: `abandon, ability, able, about, above, absent, absorb,
abstract...`

Full list: https://github.com/bitcoin/bips/blob/master/bip-0039/english.txt

## Implementation Notes

- ~100 lines of Go
- Embedded BIP-39 wordlist (~16 KB)
- Pure computation, no I/O
- Checksum uses SHA-256

```go
// Core functions
func Encode(data []byte) (string, error)
func Decode(words string) ([]byte, error)
func Validate(words string) error
```

## Limits

| Limit          | Value |
| -------------- | ----- |
| `minWords`     | 12    |
| `maxWords`     | 48    |
| `minBytes`     | 16    |
| `maxBytes`     | 64    |
| `wordlistSize` | 2048  |

## Security Considerations

- **Not encrypted**: Mnemonic is encoding, not encryption
- Anyone who hears/sees the words can decode them
- For secrets: encrypt first, then mnemonic encode
- **Checksum provides integrity, not authenticity**
- Consider environment when speaking (who's listening?)
- Paper backups should be stored securely

```bash
# For sensitive data, encrypt first:
fazt security encrypt --file secret.json | fazt mnemonic encode
# Now safe to transmit - only recipient can decrypt
```

## Comparison with Other Methods

| Method   | Bandwidth         | Hardware       | Works Through     |
| -------- | ----------------- | -------------- | ----------------- |
| Chirp    | 200 B/s           | Speaker/Mic    | Audio             |
| QR       | 2 KB/scan         | Display/Camera | Visual            |
| Mnemonic | ~3 B/s (speaking) | None           | Any human channel |

Mnemonic is slowest but most universal. Use it when:
- No shared physical space (can't use Chirp)
- No visual channel (can't use QR)
- Need to communicate through voice, text, or paper
