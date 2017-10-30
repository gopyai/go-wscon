package wscon_test

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gopyai/wscon"
)

var (
	hostPort = "localhost:8080"
	sockPath = "/"
)

func ExampleServe() {
	svr := &wscon.Server{}
	http.HandleFunc(sockPath, svr.Handler())
	isErr(http.ListenAndServe(hostPort, nil))
}

func ExampleServeTLS() {
	svr := &wscon.Server{}
	http.HandleFunc(sockPath, svr.Handler())
	isErr(http.ListenAndServeTLS(hostPort, "server.cer", "server.key", nil))
}

func ExampleConnect() {
	c := &wscon.Client{}
	isErr(c.Connect(hostPort, sockPath))
	defer c.Close()

	// connection established

	go func() {
		c.ReadLoop(func(data []byte) {
			// Dummy receive handler
		})
	}()

	go func() {
		var data []byte
		c.Write(data)
	}()

	// synchronized go routines
}

func ExampleConnectTLS() {
	c := &wscon.Client{}
	isErr(c.ConnectTLS(hostPort, sockPath))
	defer c.Close()

	// connection established

	go func() {
		c.ReadLoop(func(data []byte) {
			// Dummy receive handler
		})
	}()

	go func() {
		var data []byte
		c.Write(data)
	}()

	// synchronized go routines
}

func ExampleConnectTLSSelfSigned() {
	dialer, err := wscon.SelfSignedDialer("ca.cer")
	isErr(err)
	c := &wscon.Client{Dialer: dialer}
	isErr(c.ConnectTLS(hostPort, sockPath))
	defer c.Close()

	// connection established

	go func() {
		c.ReadLoop(func(data []byte) {
			// Dummy receive handler
		})
	}()

	go func() {
		var data []byte
		c.Write(data)
	}()

	// synchronized go routines
}

func isErr(err error) {
	if err != nil {
		fmt.Println(reflect.TypeOf(err))
		panic(err)
	}
}
