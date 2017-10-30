package wscon

import (
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

type (
	// Websocket client.
	Client struct {
		// Before initiating connection, Dialer will be checked. If it is nil, it will be set with
		// websocket.DefaultDialer. Otherwise it will be used to connect.
		Dialer *websocket.Dialer

		// The conn will be set with connection object after successfully connected using either
		// Connect() or ConnectTLS(). Thus by checking the content of conn, code can identify whether
		// the connection has been made or not (nil means not connected).
		conn *websocket.Conn

		// The wrLock is used to synchronize go routines to ensure that there is only one concurrent
		// process access WriteMessage().
		wrLock sync.Mutex

		// The rdLock is used to synchronize go routines to ensure that there is only one concurrent
		// process access ReadMessage().
		rdLock sync.Mutex
	}
)

var (
	ErrConnected    = errors.New("error: Already connected")
	ErrDisconnected = errors.New("error: No active connection")
	ErrExpectBinary = errors.New("error: Expect receive binary message")
)

// Connect to web socket server using TLS (wss://...).
func (my *Client) ConnectTLS(hostPort, path string) error {
	return my.connect("wss", hostPort, path)
}

// Connect to web socket server without using TLS (ws://...).
func (my *Client) Connect(hostPort, path string) error {
	return my.connect("ws", hostPort, path)
}

func (my *Client) connect(scheme, hostPort, path string) error {
	if my.conn != nil {
		return ErrConnected
	}
	my.defaultValues()

	u := url.URL{Scheme: scheme, Host: hostPort, Path: path}
	urlStr := u.String()
	fmt.Println("Client connect to:", urlStr)
	c, _, err := my.Dialer.Dial(urlStr, nil)
	if err != nil {
		return err
	}
	my.conn = c
	return nil
}

func (my *Client) defaultValues() {
	if my.Dialer == nil {
		my.Dialer = websocket.DefaultDialer
	}
}

// Close existing connection if it is connected.
func (my *Client) Close() {
	if my.conn == nil {
		return // already closed
	}
	my.writeCloseFrame()
	my.conn.Close()
	my.conn = nil
}

// Write binary data through web socket. It is safe for concurrent access.
func (my *Client) Write(data []byte) error {
	my.wrLock.Lock()
	defer my.wrLock.Unlock()
	if my.conn == nil {
		return ErrDisconnected
	}
	return my.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (my *Client) writeCloseFrame() error {
	my.wrLock.Lock()
	defer my.wrLock.Unlock()
	return my.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
};

// Read do forever web socket read message loop. It is safe from concurrent
// access, will return on error or after receive close frame, and will call the
// call back function whenever the message (read from server) ready.
func (my *Client) ReadLoop(receive func(data []byte)) error {
	my.rdLock.Lock()
	defer my.rdLock.Unlock()
	if my.conn == nil {
		return ErrDisconnected
	}
	for {
		msgType, msg, err := my.conn.ReadMessage()
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				my.Close()
				return nil
			}

			return err
		}
		if msgType != websocket.BinaryMessage {
			return ErrExpectBinary
		}
		receive(msg)
	}
}
