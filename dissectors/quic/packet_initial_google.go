package quic

import "gitlab.com/go-extension/bytes"

func parseGoogleLegacyInitialPacket(packet []byte, v Version) (*InitialPacket, error) {
	b := bytes.NewReader(packet)
	h, err := parseGoogleHeader(b, v)
	if err != nil {
		return nil, err
	}

	frames, err := parseGoogleFrames(b, h.Version, h.Flags.PacketNumberLen)
	if err != nil {
		return nil, err
	}

	return &InitialPacket{
		Version:      h.Version,
		PacketNumber: int64(h.PacketNumber),
		Frames:       frames,
	}, nil
}

func parseGoogleInitialPacket(packet []byte, recv bool) (*InitialPacket, error) {
	b := bytes.NewReader(packet)
	h, err := parseGoogleLongHeader(b, recv)
	if err != nil {
		return nil, err
	}
	if h.PacketType != packetTypeInitial {
		return nil, ErrNotInitialPacket
	}

	frames, err := parseGoogleFrames(b, h.Version, h.Flags.PacketNumberLen)
	if err != nil {
		return nil, err
	}

	return &InitialPacket{
		Version:      h.Version,
		PacketNumber: int64(h.PacketNumber),
		Frames:       frames,
	}, nil
}
