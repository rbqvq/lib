package quic

type packetType int

const (
	packetTypeInitial packetType = iota
	packetType0RTT
	packetTypeHandshake
	packetTypeRetry
)

func isQUIC(payload []byte) bool {
	return len(payload) >= 1200
}
