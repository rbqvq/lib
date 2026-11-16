package quic

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
	"gitlab.com/go-extension/bytes"
)

const (
	frameTypePadding = iota
	frameTypePing
	frameTypeAck
	frameTypeAckEcn
	frameTypeResetStream
	frameTypeStopSending
	frameTypeCrypto
	frameTypeNewToken
	frameTypeStream8
	frameTypeStream9
	frameTypeStreamA
	frameTypeStreamB
	frameTypeStreamC
	frameTypeStreamD
	frameTypeStreamE
	frameTypeStreamF
	frameTypeMaxData
	frameTypeMaxStreamData
	frameTypeMaxStreamsBidirectional
	frameTypeMaxStreamsUnidirectional
	frameTypeDataBlocked
	frameTypeStreamDataBlocked
	frameTypeStreamsBlockedBidirectional
	frameTypeStreamsBlockedUnidirectional
	frameTypeNewConnectionID
	frameTypeRetireConnectionID
	frameTypePathChallenge
	frameTypePathResponse
	frameTypeConnectionCloseTransport
	frameTypeConnectionCloseApplication
	frameTypeHandshakeDone
	frameTypeImmediateAck
)

const (
	frameTypeDatagram = 0x30 + iota
	frameTypeDatagramLength
)

const (
	frameTypePathAck = 0x15228c00 + iota
	frameTypePathAckEcn
	_
	_
	_
	frameTypePathAbandon
	frameTypePathStatus             // multipath-draft-05
	frameTypePathBackupAvailable    // multipath-draft-06 rename with draft-15
	frameTypePathStatusAvailable    // multipath-draft-06 rename with draft-15
	frameTypePathNewConnectionID    // multipath-draft-07
	frameTypePathRetireConnectionID // multipath-draft-07
	frameTypeMaxPaths               // multipath-draft-07
	frameTypeMaxPathID              // multipath-draft-09
	frameTypePathsBlocked           // multipath-draft-11
	frameTypePathCIDsBlocked        // multipath-draft-12
)

const (
	frameTypeImmediateAckDraft05 = 0xac //  ack-frequency-draft-05
	frameTypeAckFrequency        = 0xaf
	frameTypeTimeStamp           = 0x02F5
)

type streamFlags struct {
	Fin    bool
	Len    bool
	Offset bool
}

func parseStreamFlags(frameType byte) (flags streamFlags) {
	flags.Fin = (frameType & 0b00000001) != 0
	flags.Len = (frameType & 0b00000010) != 0
	flags.Offset = (frameType & 0b00000100) != 0
	return
}

type streamIdFlags struct {
	ClientInitiated bool
	Bidirectional   bool
}

func parseStreamIdFlags(streamId uint64) (flags streamIdFlags) {
	streamType := uint8(streamId)

	flags.ClientInitiated = (streamType & 0b00000001) == 0
	flags.Bidirectional = (streamType & 0b00000010) == 0
	return
}

func parseFrames(b *bytes.Reader) ([]FrameCrypto, error) {
	var frames []FrameCrypto

	for b.Len() > 0 {
		frameType, err := quicvarint.Read(b)
		if err != nil {
			return nil, err
		}

		switch frameType {
		case frameTypePadding:
		case frameTypePing:
		case frameTypePathAck, frameTypePathAckEcn:
			// Path ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
			fallthrough
		case frameTypeAck, frameTypeAckEcn:
			// Largest Acknowledged
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// ACK Delay
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// ACK Range Count
			ack_range_count, err := quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// First ACK Range
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			for range ack_range_count {
				// Gap
				_, err = quicvarint.Read(b)
				if err != nil {
					return nil, err
				}

				// Ack Range
				_, err = quicvarint.Read(b)
				if err != nil {
					return nil, err
				}
			}

			if frameType == frameTypeAckEcn || frameType == frameTypePathAckEcn {
				// ECT(0) Count
				_, err = quicvarint.Read(b)
				if err != nil {
					return nil, err
				}

				// ECT(1) Count
				_, err = quicvarint.Read(b)
				if err != nil {
					return nil, err
				}

				// ECN-CE Count
				_, err = quicvarint.Read(b)
				if err != nil {
					return nil, err
				}
			}
		case frameTypeResetStream:
			// Stream ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Application Error code
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Final Size
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypeStopSending:
			// Stream ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Application Error code
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypeCrypto:
			offset, err := quicvarint.Read(b) // Offset
			if err != nil {
				return nil, err
			}

			length, err := quicvarint.Read(b) // Length
			if err != nil {
				return nil, err
			}

			data, err := readFull(b, int(length)) // Data
			if err != nil {
				return nil, err
			}

			frames = append(frames, FrameCrypto{Offset: offset, Data: data})
		case frameTypeNewToken:
			length, err := quicvarint.Read(b) // Length
			if err != nil {
				return nil, err
			}

			b.Seek(int64(length), io.SeekCurrent) // Token
		case frameTypeStream8, frameTypeStream9, frameTypeStreamA, frameTypeStreamB, frameTypeStreamC, frameTypeStreamD, frameTypeStreamE, frameTypeStreamF:
			// Stream ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			flags := parseStreamFlags(byte(frameType))
			// streamFlags := parseStreamIdFlags(streamId)

			// Offset
			if flags.Offset {
				_, err = quicvarint.Read(b)
				if err != nil {
					return nil, err
				}
			}

			if flags.Len {
				length, err := quicvarint.Read(b) // Length
				if err != nil {
					return nil, err
				}

				b.Seek(int64(length), io.SeekCurrent)
			} else {
				b.Seek(0, io.SeekEnd)
			}
		case frameTypeMaxData,
			frameTypeDataBlocked:
			// Maximum Data / Stream Data Limit
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypeMaxStreamData,
			frameTypeStreamDataBlocked:
			// Stream ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Maximum Stream Data / Stream Data Limit
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypeMaxStreamsBidirectional, frameTypeMaxStreamsUnidirectional,
			frameTypeStreamsBlockedBidirectional, frameTypeStreamsBlockedUnidirectional:
			// Maximum Streams / Stream Limit
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypePathNewConnectionID:
			// Path ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
			fallthrough
		case frameTypeNewConnectionID:
			// Sequence
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Retire Prior To
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Connection ID Length
			length, err := readUint64(binary.BigEndian, b)
			if err != nil {
				return nil, err
			}

			// Connection ID
			b.Seek(int64(length), io.SeekCurrent)

			// Stateless Reset Token
			b.Seek(16, io.SeekCurrent)
		case frameTypePathRetireConnectionID:
			// Path ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
			fallthrough
		case frameTypeRetireConnectionID:
			// Sequence
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypePathChallenge, frameTypePathResponse:
			b.Seek(8, io.SeekCurrent) // Data
		case frameTypePathAbandon:
			// Path ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Error Code
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypeConnectionCloseTransport, frameTypeConnectionCloseApplication:
			// Error Code
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			if frameType == frameTypeConnectionCloseTransport {
				// Frame Type
				_, err = quicvarint.Read(b)
				if err != nil {
					return nil, err
				}
			}

			// Reason Phrase Length
			length, err := quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			b.Seek(int64(length), io.SeekCurrent) // Reason Phrase
		case frameTypeHandshakeDone:
		case frameTypeDatagram:
			b.Seek(0, io.SeekEnd) // Datagram
		case frameTypeDatagramLength:
			// Datagram Length
			length, err := quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			b.Seek(int64(length), io.SeekCurrent) // Datagram
		case frameTypeImmediateAckDraft05, frameTypeImmediateAck:
		case frameTypeAckFrequency:
			// Sequence Number
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Ack-Eliciting Threshold
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Request Max Ack Delay
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Reordering Threshold
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypeTimeStamp:
			// Time Stamp
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypePathStatus, frameTypePathBackupAvailable, frameTypePathStatusAvailable:
			// Path ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Path Status Sequence Number
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			if frameType == frameTypePathStatus {
				// Path Status
				_, err = quicvarint.Read(b)
				if err != nil {
					return nil, err
				}
			}
		case frameTypeMaxPaths, frameTypePathsBlocked:
			// Maximum Paths
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypeMaxPathID:
			// Maximum Path ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		case frameTypePathCIDsBlocked:
			// Path ID
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}

			// Next Sequence Number
			_, err = quicvarint.Read(b)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown frame type %d", frameType)
		}
	}

	return frames, nil
}
