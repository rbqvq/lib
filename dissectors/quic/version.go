package quic

type Version uint32

func (v Version) IsSupportedVersion() bool {
	switch {
	case v.isGoogle(),
		v.v1(), v.v2():
		return true
	default:
		return false
	}
}

func (v Version) isGoogle() bool {
	switch {
	case v >= VersionQ030 && v <= VersionQ039,
		v >= VersionQ040 && v <= VersionQ049,
		v == VersionQ050,
		v == VersionT050, v == VersionT051:
		return true
	default:
		return false
	}
}

func (v Version) v1() bool {
	return v == Version1 || (v >= Version1_Draft17 && v <= Version1_Draft34)
}

func (v Version) v2() bool {
	switch v {
	case Version2_Draft0, Version2_Draft1,
		Version2:
		return true
	default:
		return false
	}
}

func (v Version) TLS() bool {
	return v == VersionT050 || v == VersionT051 || v.v1() || v.v2()
}

func (v Version) isRetryLegacy() bool {
	switch v {
	case VersionQ050, VersionT050, VersionT051,
		Version1_Draft25, Version1_Draft26, Version1_Draft27, Version1_Draft28, Version1_Draft29, Version1_Draft30,
		Version1_Draft31, Version1_Draft32, Version1_Draft33, Version1_Draft34,
		Version1,
		Version2_Draft0, Version2_Draft1,
		Version2:
		return false
	default:
		return true
	}
}

func ParseVersion(packet []byte) (Version, error) {
	if !isQUIC(packet) {
		return 0, ErrNotQUIC
	}

	typeByte := packet[0]
	switch {
	case isHeader(typeByte) && isLongHeader(typeByte):
		return Version(beUint32(packet[1:5])), nil
	case isGoogleHeaderWithVersion(typeByte):
		return Version(beUint32(packet[9:13])), nil
	default:
		return 0, ErrNotQUIC
	}
}
