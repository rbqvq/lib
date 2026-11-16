package http

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"syscall"

	libhttp "gitlab.com/CoiaPrant/lib/net/http"
	"golang.org/x/net/proxy"
)

func init() {
	proxy.RegisterDialerType("http", NewDialer)
	proxy.RegisterDialerType("https", NewDialer)
}

type proxyDialer struct {
	raddr     string
	headers   http.Header
	tlsConfig *tls.Config

	dialer proxy.Dialer
}

func NewDialer(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	if u == nil {
		return nil, syscall.EINVAL
	}
	if forward == nil {
		forward = proxy.Direct
	}

	if u.Host == "" {
		return nil, fmt.Errorf("lib/net/proxy/http: missing address in URL")
	}

	d := &proxyDialer{
		raddr:   u.Host,
		headers: make(http.Header),
		dialer:  forward,
	}
	switch u.Scheme {
	case "http":
	case "https":
		insecure, _ := strconv.ParseBool(u.Query().Get("insecure"))
		d.tlsConfig = &tls.Config{
			ServerName:         u.Hostname(),
			InsecureSkipVerify: insecure,
		}
	default:
		return nil, fmt.Errorf("lib/net/proxy/http: invalid scheme in URL")
	}

	if u.User != nil {
		username := u.User.Username()
		password, _ := u.User.Password()
		d.headers.Set("Proxy-Authorization", "Basic "+libhttp.BasicAuth(username, password))
	}

	return d, nil
}

func (d *proxyDialer) addTLS(ctx context.Context, c net.Conn) (net.Conn, error) {
	if d.tlsConfig == nil {
		return c, nil
	}

	conn := tls.Client(c, d.tlsConfig)
	if err := conn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (d *proxyDialer) Dial(network, address string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, syscall.EAFNOSUPPORT
	}

	c, err := d.dialer.Dial("tcp", d.raddr)
	if err != nil {
		return nil, err
	}

	conn, err := d.addTLS(context.Background(), c)
	if err != nil {
		c.Close()
		return nil, err
	}

	req := &http.Request{
		Method:     "CONNECT",
		Host:       address,
		URL:        &url.URL{Opaque: address},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}
	maps.Copy(req.Header, d.headers)

	err = req.Write(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	br := bufio.NewReader(conn)
	response, err := http.ReadResponse(br, req)
	if err != nil {
		conn.Close()
		return nil, err
	}
	response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		conn.Close()
		return nil, fmt.Errorf("lib/net/proxy/http: proxy response code: %d", response.StatusCode)
	}

	if response.ContentLength != -1 || len(response.TransferEncoding) != 0 {
		conn.Close()
		return nil, errors.New("lib/net/proxy/http: proxy returned unexpected body")
	}

	if br.Buffered() == 0 {
		return conn, nil
	}

	return &netConn{Conn: conn, Reader: br}, nil
}

func (d *proxyDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	dialer, ok := d.dialer.(proxy.ContextDialer)
	if !ok {
		return d.Dial(network, address)
	}

	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, syscall.EAFNOSUPPORT
	}

	c, err := dialer.DialContext(ctx, "tcp", d.raddr)
	if err != nil {
		return nil, err
	}

	conn, err := d.addTLS(ctx, c)
	if err != nil {
		c.Close()
		return nil, err
	}

	req := (&http.Request{
		Method:     "CONNECT",
		Host:       address,
		URL:        &url.URL{Opaque: address},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}).WithContext(ctx)
	maps.Copy(req.Header, d.headers)

	err = req.Write(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	br := bufio.NewReader(conn)
	response, err := http.ReadResponse(br, req)
	if err != nil {
		conn.Close()
		return nil, err
	}
	response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		conn.Close()
		return nil, fmt.Errorf("lib/net/proxy/http: proxy response code: %d", response.StatusCode)
	}

	if response.ContentLength != -1 || len(response.TransferEncoding) != 0 {
		conn.Close()
		return nil, errors.New("lib/net/proxy/http: proxy returned unexpected body")
	}

	if br.Buffered() == 0 {
		return conn, nil
	}

	return &netConn{Conn: conn, Reader: br}, nil
}
