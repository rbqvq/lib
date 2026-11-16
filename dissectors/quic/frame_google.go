package quic

import (
	"encoding/binary"
	"fmt"
	"io"

	"gitlab.com/go-extension/bytes"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	frameTypeGooglePadding = iota
	frameTypeGoogleResetStream
	frameTypeGoogleConnectionClose
	frameTypeGoogleGoaway
	frameTypeGoogleWindowUpdate
	frameTypeGoogleBlocked
	frameTypeGoogleStopWaiting
	frameTypeGooglePing
	frameTypeGoogleCrypto
)

type googleStreamFlags struct {
	Fin       bool
	DataLen   bool
	OffsetLen int
	StreamLen int
}

func isGoogleStream(frameType byte) bool {
	return (frameType & 0b10000000) != 0
}

func parseGoogleStreamFlags(frameType byte) (flags googleStreamFlags) {
	flags.Fin = (frameType & 0b01000000) != 0
	flags.DataLen = (frameType & 0b00100000) != 0

	if offsetLength := (frameType & 0b00011100) >> 2; offsetLength != 0 {
		flags.OffsetLen = int(offsetLength) + 1
	}

	flags.StreamLen = int((frameType & 0b00000011) + 1)
	return
}

type googleAckFlags struct {
	NACK               bool
	UnusedOrTruncated  bool // Unused since Q034
	LargestObservedLen int
	MissingPacketLen   int
}

func isGoogleAck(frameType byte) bool {
	return (frameType & 0b01000000) != 0
}

func parseGoogleAckFlags(frameType byte) (flags googleAckFlags) {
	flags.NACK = (frameType & 0b00100000) != 0
	flags.UnusedOrTruncated = (frameType & 0b00010000) != 0

	switch (frameType & 0b00001100) >> 2 {
	case 0:
		flags.LargestObservedLen = 1
	case 1:
		flags.LargestObservedLen = 2
	case 2:
		flags.LargestObservedLen = 4
	case 3:
		flags.LargestObservedLen = 6
	}

	switch frameType & 0b00000011 {
	case 0:
		flags.MissingPacketLen = 1
	case 1:
		flags.MissingPacketLen = 2
	case 2:
		flags.MissingPacketLen = 4
	case 3:
		flags.MissingPacketLen = 6
	}
	return
}

func parseGoogleFrames(b *bytes.Reader, v Version, packetNumberLen int64) ([]FrameCrypto, error) {
	var frames []FrameCrypto

	var byteOrder binary.ByteOrder = binary.BigEndian
	if v < VersionQ039 {
		byteOrder = binary.LittleEndian
	}

	for b.Len() > 0 {
		frameType, err := b.ReadByte()
		if err != nil {
			return nil, err
		}

		switch frameType {
		case frameTypeGooglePadding:
		case frameTypeGoogleResetStream:
			b.Seek(4, io.SeekCurrent) // Stream ID
			b.Seek(8, io.SeekCurrent) // Byte Offset
			b.Seek(4, io.SeekCurrent) // Error Code
		case frameTypeGoogleConnectionClose:
			b.Seek(4, io.SeekCurrent) // Error Code

			length, err := readUint16(byteOrder, b) // Reason Phrase Length
			if err != nil {
				return nil, err
			}

			b.Seek(int64(length), io.SeekCurrent) // Reason Phrase
		case frameTypeGoogleGoaway:
			b.Seek(4, io.SeekCurrent) // Error Code
			b.Seek(4, io.SeekCurrent) // Last Good Stream ID

			length, err := readUint16(byteOrder, b) // Reason Phrase Length
			if err != nil {
				return nil, err
			}

			b.Seek(int64(length), io.SeekCurrent) // Reason Phrase
		case frameTypeGoogleWindowUpdate:
			b.Seek(4, io.SeekCurrent) // Stream ID
			b.Seek(8, io.SeekCurrent) // Byte Offset
		case frameTypeGoogleBlocked:
			b.Seek(4, io.SeekCurrent) // Stream ID
		case frameTypeGoogleStopWaiting:
			if v < VersionQ034 {
				b.Seek(1, io.SeekCurrent) // Send Entropy
			}
			b.Seek(packetNumberLen, io.SeekCurrent) // Least Unacked Delta
		case frameTypeGooglePing:
		case frameTypeGoogleCrypto:
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
		default:
			switch {
			case isGoogleStream(frameType):
				flags := parseGoogleStreamFlags(frameType)

				// Stream ID
				streamId, err := readUint(byteOrder, b, flags.StreamLen)
				if err != nil {
					return nil, err
				}

				// Offset
				offset, err := readUint(byteOrder, b, flags.OffsetLen)
				if err != nil {
					return nil, err
				}

				// Length
				if !flags.DataLen {
					continue
				}

				length, err := readUint16(byteOrder, b)
				if err != nil {
					return nil, err
				}

				if streamId != 1 { // Reserved Google QUIC (handshake, crypto, config updates...)
					b.Seek(int64(length), io.SeekCurrent)
					continue
				}

				data, err := readFull(b, int(length)) // Data
				if err != nil {
					return nil, err
				}

				frames = append(frames, FrameCrypto{Offset: offset, Data: data})
			case isGoogleAck(frameType):
				flags := parseGoogleAckFlags(frameType)
				if v < VersionQ034 {
					// Received Entropy
					b.Seek(1, io.SeekCurrent)

					// Largest Observed
					b.Seek(int64(flags.LargestObservedLen), io.SeekCurrent)

					// Ack Delay time
					b.Seek(2, io.SeekCurrent)

					// Timestamp
					num_timestamp, err := b.ReadByte()
					if err != nil {
						return nil, err
					}
					if num_timestamp > 0 {
						// Delta Largest Observed
						b.Seek(1, io.SeekCurrent)

						// First Timestamp
						b.Seek(4, io.SeekCurrent)

						// Num Timestamp (-1) x (Delta Largest Observed + Time Since Previous Timestamp)
						b.Seek(int64(num_timestamp-1)*(1+2), io.SeekCurrent)

						if flags.NACK {
							// Ack Range
							num_ranges, err := b.ReadByte()
							if err != nil {
								return nil, err
							}

							if num_ranges > 0 {
								// Num Ranges x (Missing Packet Number Delta + Range Length)
								b.Seek(int64(num_ranges)*int64(flags.MissingPacketLen+1), io.SeekCurrent)
							}

							// Ack Revived
							num_revived, err := b.ReadByte()
							if err != nil {
								return nil, err
							}

							if num_revived > 0 {
								// Num Revived x (Revived Packet Number)
								b.Seek(int64(num_revived)*int64(flags.LargestObservedLen), io.SeekCurrent)
							}
						}
					}
					continue
				}

				// Largest Acked
				b.Seek(int64(flags.LargestObservedLen), io.SeekCurrent)

				// Largest Acked Delta Time
				b.Seek(2, io.SeekCurrent)

				// Ack Block
				var num_blocks uint8
				if flags.NACK {
					num_blocks, err = b.ReadByte()
					if err != nil {
						return nil, err
					}
				}

				// First ACK block
				b.Seek(int64(flags.MissingPacketLen), io.SeekCurrent)
				if num_blocks > 0 {
					// Num Blocks x (Gap + Ack Block)
					b.Seek(int64(num_blocks)*int64(1+flags.MissingPacketLen), io.SeekCurrent)
				}

				// Timestamp
				num_timestamp, err := b.ReadByte()
				if err != nil {
					return nil, err
				}
				if num_timestamp > 0 {
					// Delta Largest Acked
					b.Seek(1, io.SeekCurrent)

					// Time Since Largest Acked
					b.Seek(4, io.SeekCurrent)

					// Num Timestamp x (Delta Largest Acked + Time Since Previous Timestamp)
					b.Seek(int64(num_timestamp-1)*(1+2), io.SeekCurrent)
				}
			default:
				return nil, fmt.Errorf("unknown google frame type %d", frameType)
			}
		}
	}

	return frames, nil
}
