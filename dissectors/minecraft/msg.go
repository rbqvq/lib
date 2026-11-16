package minecraft

import "io"

type Message interface {
	Format(version uint32) ([]byte, error)
}

func WriteToMsg(w io.Writer, version uint32, m Message) (int64, error) {
	data, err := m.Format(version)
	if err != nil {
		return 0, err
	}

	return (&Packet{Data: data}).WriteTo(w)
}
