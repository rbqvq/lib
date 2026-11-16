package quic

import "sort"

type FrameCrypto struct {
	Offset uint64
	Data   []byte
}

// AssembleCryptoFrames assembles multiple crypto frames into a single slice.
// It returns false if the frames cannot be assembled. This can happen if the frames are not contiguous.
func AssembleCryptoFrames(frames []FrameCrypto) ([]byte, bool) {
	if len(frames) == 0 {
		return nil, true
	}

	// sort the frames by offset
	sort.Slice(frames, func(i, j int) bool { return frames[i].Offset < frames[j].Offset })

	// check offset
	if frames[0].Offset != 0 {
		return nil, false
	}

	if len(frames) == 1 {
		return frames[0].Data, true
	}

	// check if the frames are contiguous
	var offset uint64
	assembled := make([]FrameCrypto, 0, len(frames))
	for _, frame := range frames {
		if frame.Offset > offset {
			return nil, false
		}

		length := uint64(len(frame.Data))
		end := frame.Offset + length
		if frame.Offset == offset {
			assembled = append(assembled, frame)
			offset = end
			continue
		}

		if offset >= end {
			// Ignore already exists frames
			continue
		}

		trim := length - (end - offset)
		assembled = append(assembled, FrameCrypto{
			Offset: offset,
			Data:   frame.Data[trim:],
		})
		offset = end
	}

	// concatenate the frames
	data := make([]byte, offset)
	for _, frame := range assembled {
		copy(data[frame.Offset:], frame.Data)
	}
	return data, true
}
