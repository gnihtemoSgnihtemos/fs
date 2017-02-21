package ftp

import (
	"bufio"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"testing"
	"time"
)

const server = `220 Service ready for new user.
221 Closing control connection.
`

type fakeConn struct {
	io.ReadWriter
	readDeadline *deadline
}

type fakeClock struct{ now time.Time }

type deadline struct{ time time.Time }

func (c fakeConn) Close() error                     { return nil }
func (c fakeConn) LocalAddr() net.Addr              { return nil }
func (c fakeConn) RemoteAddr() net.Addr             { return nil }
func (c fakeConn) SetDeadline(time.Time) error      { return nil }
func (c fakeConn) SetWriteDeadline(time.Time) error { return nil }

func (c fakeConn) SetReadDeadline(t time.Time) error {
	c.readDeadline.time = t
	return nil
}

func (c fakeClock) Now() time.Time { return c.now }

func TestReadDeadline(t *testing.T) {
	conn := fakeConn{readDeadline: &deadline{}}
	w := bufio.NewWriter(ioutil.Discard)
	conn.ReadWriter = bufio.NewReadWriter(bufio.NewReader(strings.NewReader(server)), w)

	clock := fakeClock{now: time.Date(2017, 2, 1, 16, 35, 0, 0, time.UTC)}
	initialTimeout := time.Second * 30

	client, err := newClient(conn, initialTimeout, clock)
	if err != nil {
		t.Fatal(err)
	}
	if want := clock.Now().Add(initialTimeout); !conn.readDeadline.time.Equal(want) {
		t.Errorf("want read deadline %s, got %s", want, conn.readDeadline.time)
	}

	client.ReadTimeout = time.Minute
	if _, _, err := client.Cmd(221, "QUIT"); err != nil {
		t.Fatal(err)
	}
	if want := clock.Now().Add(client.ReadTimeout); !conn.readDeadline.time.Equal(want) {
		t.Errorf("want read deadline %s, got %s", want, conn.readDeadline.time)
	}
}
