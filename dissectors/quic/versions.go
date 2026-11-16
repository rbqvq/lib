package quic

const (
	VersionNegotiation Version = 0
)

// Google QUIC
const (
	VersionQ030 Version = 0x51303330 + iota
	VersionQ031
	VersionQ032
	VersionQ033
	VersionQ034
	VersionQ035
	VersionQ036
	VersionQ037
	VersionQ038
	VersionQ039
)

const (
	VersionQ040 Version = 0x51303430 + iota
	VersionQ041
	VersionQ042
	VersionQ043
	VersionQ044
	VersionQ045
	VersionQ046
	VersionQ047
	VersionQ048
	VersionQ049
)

const (
	VersionQ050 Version = 0x51303530
)

// Google QUIC with TLS
const (
	VersionT050 Version = 0x54303530
	VersionT051 Version = 0x54303531
)

// QUIC Draft v1
const (
	Version1_Draft17 Version = 0xff000011 + iota
	Version1_Draft18
	Version1_Draft19
	Version1_Draft20
	Version1_Draft21
	Version1_Draft22
	Version1_Draft23
	Version1_Draft24
	Version1_Draft25
	Version1_Draft26
	Version1_Draft27
	Version1_Draft28
	Version1_Draft29
	Version1_Draft30
	Version1_Draft31
	Version1_Draft32
	Version1_Draft33
	Version1_Draft34
)

// QUIC v1 (RFC 9001)
const (
	Version1 Version = 0x1
)

// QUIC v2 (RFC 9369)
const (
	Version2_Draft0 Version = 0xff020000
	Version2_Draft1 Version = 0x709a50c4
	Version2        Version = 0x6b3343cf
)
