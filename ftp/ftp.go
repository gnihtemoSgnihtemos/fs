package ftp

import (
	"crypto/tls"
	"net"
	"net/textproto"
	"time"
)

type Client struct {
	conn net.Conn
	text *textproto.Conn
}

func newClient(conn net.Conn) (*Client, error) {
	client := &Client{
		conn: conn,
		text: textproto.NewConn(conn),
	}
	// Read multiline 2xx responses sent by server before login
	_, _, err := client.text.ReadResponse(2)
	return client, err
}

func Dial(network, addr string) (*Client, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return newClient(conn)
}

func DialTimeout(network, addr string, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	return newClient(conn)
}

func (c *Client) Close() error {
	return c.text.Close()
}

func (c *Client) Cmd(expectCode int, format string, args ...interface{}) (int, string, error) {
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
	code, message, err := c.Cmd(0, "USER %s", user)
	if err != nil || code == 230 { // 230 = User logged in, proceed.
		return err
	}
	if code != 331 { // 331 = User name okay, need password.
		return &textproto.Error{Code: code, Msg: message}
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
