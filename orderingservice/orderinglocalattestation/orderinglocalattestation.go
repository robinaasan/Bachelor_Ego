package orderinglocalattestation

// This code is inspired from a sample created by Edgeless systems regarding local attestation
// link: https://github.com/edgelesssys/ego/tree/master/samples
import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

func CreateServerCertificate() ([]byte, crypto.PrivateKey) {
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

func CreateClientCertificate(pubKey []byte, signerCert []byte, signerPrivKey crypto.PrivateKey) []byte {
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

func GetQueryArg(w http.ResponseWriter, r *http.Request, name string) []byte {
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
