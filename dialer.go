package wscon

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/gorilla/websocket"
)

func SelfSignedDialer(caCertName string) (*websocket.Dialer, error) {
	caCer, err := loadCert(caCertName)
	if err != nil {
		return nil, err
	}
	root := x509.NewCertPool()
	root.AddCert(caCer)
	return &websocket.Dialer{TLSClientConfig: &tls.Config{RootCAs: root}}, nil
}

func loadCert(certName string) (*x509.Certificate, error) {
	b, err := ioutil.ReadFile(certName)
	if err != nil {
		return nil, err
	}
	blk, _ := pem.Decode(b)
	return x509.ParseCertificate(blk.Bytes)
}
