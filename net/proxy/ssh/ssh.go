package ssh

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"weak"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

const clientVersion = "SSH-2.0-OpenSSH_10.0p2 Debian-7+deb13u1"

func init() {
	proxy.RegisterDialerType("ssh", NewDialer)
}

type proxyDialer struct {
	raddr  string
	dialer proxy.Dialer

	client      *atomic.Pointer[ssh.Client]
	user        string
	authMethods []ssh.AuthMethod
}

func maintainClient(d weak.Pointer[proxyDialer], client *ssh.Client) {
	for {
		client.Wait()
		client.Close()

	obtainClient:
		d := d.Value()
		if d == nil {
			return
		}

		c, err := d.obtainClient()
		if err != nil {
			d = nil // release reference to proxyDialer
			time.Sleep(time.Second)
			goto obtainClient
		}

		client = c
		d.client.Store(client)
		d = nil // release reference to proxyDialer
	}
}

func freeClient(client *atomic.Pointer[ssh.Client]) {
	client.Load().Close()
}

func NewDialer(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	if u == nil {
		return nil, syscall.EINVAL
	}
	if forward == nil {
		forward = proxy.Direct
	}

	if u.Scheme != "ssh" {
		return nil, fmt.Errorf("lib/net/proxy/ssh: invalid scheme in URL")
	}

	if u.Host == "" {
		return nil, fmt.Errorf("lib/net/proxy/ssh: missing address in URL")
	}

	if u.User == nil {
		return nil, fmt.Errorf("lib/net/proxy/ssh: missing user info in URL")
	}

	d := &proxyDialer{
		raddr:       u.Host,
		dialer:      forward,
		client:      new(atomic.Pointer[ssh.Client]),
		user:        u.User.Username(),
		authMethods: make([]ssh.AuthMethod, 0, 2),
	}

	if password, ok := u.User.Password(); ok {
		d.authMethods = append(d.authMethods, ssh.Password(password))
	}

	if key := u.Query().Get("key"); key != "" {
		privkey, err := os.ReadFile(key)
		if err != nil {
			return nil, err
		}

		signer, err := ssh.ParsePrivateKey(privkey)
		if err != nil {
			return nil, err
		}

		d.authMethods = append(d.authMethods, ssh.PublicKeys(signer))
	}

	client, err := d.obtainClient()
	if err != nil {
		return nil, err
	}
	d.client.Store(client)

	go maintainClient(weak.Make(d), client)
	runtime.AddCleanup(d, freeClient, d.client)

	return d, nil
}

func (d *proxyDialer) obtainClient() (*ssh.Client, error) {
	c, err := d.dialer.Dial("tcp", d.raddr)
	if err != nil {
		return nil, err
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(c, d.raddr, &ssh.ClientConfig{
		ClientVersion:   clientVersion,
		Auth:            d.authMethods,
		User:            d.user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		c.Close()
		return nil, err
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

func (d *proxyDialer) Dial(network, address string) (net.Conn, error) {
	return d.client.Load().Dial(network, address)
}

func (d *proxyDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.client.Load().DialContext(ctx, network, address)
}
