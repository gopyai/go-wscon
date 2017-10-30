package wscon

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type (
	Server struct {
		// Callback handler. Will be called when new client is connected.
		OnConnect func(clientId string)

		// Callback handler. Will be called when there is data available (read from client).
		OnReceive func(clientId string, data []byte)

		// It is a map containing client connections object which identified by (map key) client id.
		clientConn struct {
			sync.Mutex
			m map[string]*websocket.Conn
		}
	}
)

var (
	upgrader     = websocket.Upgrader{}
	ErrInvalidId = errors.New("error: Invalid client id")
)

func (my *Server) Handler() http.HandlerFunc {
	my.clientConn.m = make(map[string]*websocket.Conn)
	return my.handler
}

func (my *Server) handler(w http.ResponseWriter, r *http.Request) {
	clientId := strings.Split(r.RemoteAddr, ":")[0]

	my.clientConn.Lock()
	if _, ok := my.clientConn.m[clientId]; ok {
		my.clientConn.Unlock()
		http.Error(w, "Duplicate ID", 400)
		return
	}
	my.clientConn.Unlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer conn.Close()

	if my.OnConnect != nil {
		my.OnConnect(clientId)
	}

	my.clientConn.Lock()
	my.clientConn.m[clientId] = conn
	my.clientConn.Unlock()

	my.wsHandler(clientId, conn)

	my.clientConn.Lock()
	delete(my.clientConn.m, clientId)
	my.clientConn.Unlock()
}

func (my *Server) wsHandler(clientId string, conn *websocket.Conn) {
	fmt.Println("Accept web socket connection from", clientId)
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				fmt.Println("Receive close frame from:", clientId)
				return // Immediately close connection when receive close frame
			}

			// Error
			fmt.Println("Error type:", reflect.TypeOf(err), "error:", err)
			return
		}
		if msgType != websocket.BinaryMessage {
			fmt.Println("Error: Read message got non binary message")
			return
		}
		if my.OnReceive != nil {
			my.OnReceive(clientId, msg)
		}
	}
}

// Write checks client id and write the data to client specified. It is safe
// from concurrent access.
func (my *Server) Write(clientId string, data []byte) error {
	my.clientConn.Lock()
	defer my.clientConn.Unlock()
	conn, ok := my.clientConn.m[clientId]
	if !ok {
		return ErrInvalidId
	}
	return conn.WriteMessage(websocket.BinaryMessage, data)
}
