# Ideas

- Support multiple DBs
- Implement gpg
- have PHP server / CGI?
- add sensor layer?
  - so that pulse will also get sensor data when available:
    image/video/audio/anything else?
  - rationale: VLMs/multi-modal models are getting popular; i think locally
    runnable ones will come soon; so having an awareness system with a sensor
    layer will make fazt a portable brain for any fazt-runnable machine

- make all data formats JSON
- so that we can do custom renders later, if needed
- making it fully JSON with validated types might make the system more robust?
- plus I think in future JSON to UI will be solved

## Wazero for any language
- what if the runtime can inject a ruby/python library like fazt js library?

- is it sensible to move js library out to a separate repo (may be not now, but
  later?)

- there is a project, I have cloned in ~/Projects/go-periph-io ; may be review
  it in connection with the sensor section?

- fazt as an mcp client, hence enabling:
  - browser control
  - computer control
  - mobile device control

- money as resource
  - what if we introduce a concept of "value" (or money) into fazt
  - so that each activity need money to operate
  - it need not be real
  - but something like a virtual currency internally that can be used as a
    baseline 
  - rationale: we can allote some money to fazt and set some costs for different
    activities and see how interesting optimisations/patterns emerge
  - we can even denote this as BTC; some super tiny amounts
  - why: so eventually we can think of moving this into some swarm that actually
    has a monetary primitive
  - can spend, expand, collect resources, provide value, and autonomously exist
  - i know the idea is far fetching & possibly too vague and crazy; but i wanted
    to atleast think of these lines to create some notes at the minimum

- lite versions
  - bitcoin? primitives?
  - bittorrent primitives?
  - lite gpg? without full bloat
