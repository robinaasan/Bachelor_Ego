package localattestation

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/edgelesssys/ego/enclave"
)

func newAttestServer(cert []byte, privKey crypto.PrivateKey) *http.Server {
	certHash := sha256.Sum256(cert)
	mux := http.NewServeMux()

	// Returns the server certificate.
	mux.HandleFunc("/cert", func(w http.ResponseWriter, r *http.Request) { w.Write(cert) })

	// Returns a local report including the server certificate's hash for the given target report.
	mux.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		targetReport := getQueryArg(w, r, "target")
		if targetReport == nil {
			return
		}
		report, err := enclave.GetLocalReport(certHash[:], targetReport)
		if err != nil {
			http.Error(w, fmt.Sprintf("GetLocalReport: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write(report)
	})

	// Returns a client certificate for the given pubkey.
	// The given report ensures that only verified enclaves can get certificates for their pubkeys.
	mux.HandleFunc("/client", func(w http.ResponseWriter, r *http.Request) {
		pubKey := getQueryArg(w, r, "pubkey")
		if pubKey == nil {
			return
		}
		report := getQueryArg(w, r, "report")
		if report == nil {
			return
		}
		if err := verifyReport(report, pubKey); err != nil {
			http.Error(w, fmt.Sprintf("verifyReport: %v", err), http.StatusBadRequest)
			return
		}
		w.Write(createClientCertificate(pubKey, cert, privKey))
	})

	return &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
}


func newSecureServer(cert []byte, privKey crypto.PrivateKey) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "pong") })

	// use server certificate also as client CA
	parsedCert, _ := x509.ParseCertificate(cert)
	clientCAs := x509.NewCertPool()
	clientCAs.AddCert(parsedCert)

	return &http.Server{
		Addr:    "localhost:8081",
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{cert},
					PrivateKey:  privKey,
				},
			},
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  clientCAs,
		},
	}
}


func createServerCertificate() ([]byte, crypto.PrivateKey) {
	template := &x509.Certificate{
		SerialNumber:          &big.Int{},
		Subject:               pkix.Name{CommonName: "server"},
		NotAfter:              time.Now().Add(time.Hour),
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
	}
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert, _ := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	return cert, priv
}


func createClientCertificate(pubKey []byte, signerCert []byte, signerPrivKey crypto.PrivateKey) []byte {
	template := &x509.Certificate{
		SerialNumber: &big.Int{},
		Subject:      pkix.Name{CommonName: "client"},
		NotAfter:     time.Now().Add(time.Hour),
	}
	parsedPubKey, _ := x509.ParsePKCS1PublicKey(pubKey)
	parsedSignerCert, _ := x509.ParseCertificate(signerCert)
	cert, _ := x509.CreateCertificate(rand.Reader, template, parsedSignerCert, parsedPubKey, signerPrivKey)
	return cert
}

func getQueryArg(w http.ResponseWriter, r *http.Request, name string) []byte {
	values := r.URL.Query()[name]
	if len(values) == 0 {
		http.Error(w, fmt.Sprintf("query argument not found: %v", name), http.StatusBadRequest)
		return nil
	}
	result, err := base64.URLEncoding.DecodeString(values[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("decoding query argument '%v' failed: %v", name, err), http.StatusBadRequest)
		return nil
	}
	return result
}
