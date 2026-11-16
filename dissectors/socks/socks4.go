package socks

import "gitlab.com/go-extension/bytes"

func parseV4(r *bytes.Reader) (Command, error) {
	v, err := readVersion(r)
	if err != nil {
		return 0, err
	}
	if v != Version4 {
		return 0, ErrNotSOCKS
	}

	cmd, err := readCommand(r)
	if err != nil {
		return 0, err
	}
	if !cmd.ValidV4() {
		return 0, ErrNotSOCKS
	}
	return cmd, nil
}

func ParseRequestV4(packet []byte) (*RequestV4, error) {
	b := bytes.NewReader(packet)

	cmd, err := parseV4(b)
	if err != nil {
		return nil, err
	}

	req, err := parseRequestBodyV4(b)
	if err != nil {
		return nil, err
	}
	req.Command = cmd
	return req, nil
}
