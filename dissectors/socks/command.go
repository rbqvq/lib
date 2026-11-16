package socks

type Command uint8

const (
	CmdTcpConnect Command = iota
	CmdTcpBind
	CmdUdpAssociate
)

func (cmd Command) ValidV4() bool {
	switch cmd {
	case CmdTcpConnect, CmdTcpBind:
		return true
	default:
		return false
	}
}

func (cmd Command) ValidV5() bool {
	switch cmd {
	case CmdTcpConnect, CmdTcpBind, CmdUdpAssociate:
		return true
	default:
		return false
	}
}
