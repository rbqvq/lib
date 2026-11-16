# QUIC

A collection of QUIC protocol dissectors and utilities.

## Features

- QUIC Initial packet parsing
- QUIC ClientHello (TLS over QUIC) parsing
- gQUIC packet parsing
- Zero-copy parsing where possible
- Pure Go implementation
- Standard library–first design with minimal dependencies

## References

This implementation was developed with reference to the following public
projects and specifications.

### Wireshark

- `epan/dissectors/packet-quic.c`
  https://github.com/wireshark/wireshark/commit/a2eecbd2b561c9bcc7e300cbee12707f587f4a92

- `epan/dissectors/packet-gquic.c`
  https://github.com/wireshark/wireshark/commit/a3c3a7cd3819be65db9771a89de6725b4d7c8f2f

### IETF RFCs

- RFC 8999 — Version-Independent Properties of QUIC
- RFC 9000 — QUIC: A UDP-Based Multiplexed and Secure Transport
- RFC 9001 — Using TLS to Secure QUIC
- RFC 9002 — QUIC Loss Detection and Congestion Control

### Drafts

- Drafts and historical specifications for Google QUIC (gQUIC), where applicable.