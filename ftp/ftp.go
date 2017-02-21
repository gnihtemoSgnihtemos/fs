package ftp

import (
	"crypto/tls"
	"net"
	"net/textproto"
	"time"
)

type clock interface {
	Now() time.Time
}

type realClock struct{}

func (r realClock) Now() time.Time { return time.Now() }

type Client struct {
	conn        net.Conn
	text        *textproto.Conn
	clock       clock
	ReadTimeout time.Duration
}

func newClient(conn net.Conn, timeout time.Duration, clock clock) (*Client, error) {
	c := &Client{
		conn:  conn,
		text:  textproto.NewConn(conn),
		clock: clock,
	}
	// Read multiline 220 responses sent by server before login
	c.setReadTimeout(timeout)
	_, _, err := c.text.ReadResponse(220)
	return c, err
}

func NewClient(conn net.Conn, timeout time.Duration) (*Client, error) {
	return newClient(conn, timeout, realClock{})
}

func Dial(network, addr string) (*Client, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return NewClient(conn, 0)
}

func DialTimeout(network, addr string, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	return NewClient(conn, timeout)
}

func (c *Client) setReadTimeout(timeout time.Duration) {
	if timeout == 0 {
		return
	}
	deadline := c.clock.Now().Add(timeout)
	c.conn.SetReadDeadline(deadline)
}

func (c *Client) Close() error {
	return c.text.Close()
}

func (c *Client) Cmd(expectCode int, format string, args ...interface{}) (int, string, error) {
	c.setReadTimeout(c.ReadTimeout)
	if err := c.text.PrintfLine(format, args...); err != nil {
		return 0, "", err
	}
	return c.text.ReadResponse(expectCode)
}

func (c *Client) AuthTLS(config *tls.Config) error {
	if _, _, err := c.Cmd(234, "AUTH TLS"); err != nil {
		return err
	}
	c.conn = tls.Client(c.conn, config)
	c.text = textproto.NewConn(c.conn)
	return nil
}

func (c *Client) Stat(args string) (string, error) {
	_, message, err := c.Cmd(213, "STAT %s", args)
	return message, err
}

func (c *Client) Quit() error {
	_, _, err := c.Cmd(221, "QUIT")
	if err != nil {
		return err
	}
	return c.Close()
}

func (c *Client) Login(user, pass string) error {
	_, _, err := c.Cmd(331, "USER %s", user)
	if err != nil {
		return err
	}
	_, _, err = c.Cmd(230, "PASS %s", pass)
	return err
}

func (c *Client) LoginWithTLS(config *tls.Config, user, pass string) error {
	if err := c.AuthTLS(config); err != nil {
		return err
	}
	if err := c.Login(user, pass); err != nil {
		return err
	}
	if _, _, err := c.Cmd(200, "PBSZ 0"); err != nil {
		return err
	}
	if _, _, err := c.Cmd(200, "PROT P"); err != nil {
		return err
	}
	return nil
}
