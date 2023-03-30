package runtimelocalattestation

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// func main() {
// 	const attestURL = "http://localhost:8087"
// 	const secureURL = "https://localhost:8088"

// 	// create client keys
// 	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
// 	pubKey := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)

// 	// get server certificate over insecure channel
// 	serverCert := httpGet(nil, attestURL+"/cert")

// 	// get the server's report targeted at this client
// 	clientInfoReport, err := enclave.GetLocalReport(nil, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	serverReport := httpGet(nil, attestURL+"/report", makeArg("target", clientInfoReport))

// 	// verify server certificate using the server's report
// 	if err := verifyreport.VerifyReport(serverReport, serverCert); err != nil {
// 		panic(err)
// 	}

// 	// request a client certificate from the server
// 	pubKeyHash := sha256.Sum256(pubKey)
// 	clientReport, err := enclave.GetLocalReport(pubKeyHash[:], serverReport)
// 	if err != nil {
// 		panic(err)
// 	}
// 	clientCert := httpGet(nil, attestURL+"/client", makeArg("pubkey", pubKey), makeArg("report", clientReport))

// 	// create mutual TLS config
// 	tlsConfig := &tls.Config{
// 		Certificates: []tls.Certificate{
// 			{
// 				Certificate: [][]byte{clientCert},
// 				PrivateKey:  privKey,
// 			},
// 		},
// 		RootCAs: x509.NewCertPool(),
// 	}
// 	parsedServerCert, _ := x509.ParseCertificate(serverCert)
// 	tlsConfig.RootCAs.AddCert(parsedServerCert)

// 	// use the established secure channel
// 	resp := httpGet(tlsConfig, secureURL+"/ping")
// 	fmt.Printf("server responded: %s\n", resp)
// }

func HttpGet(tlsConfig *tls.Config, url string, args ...string) []byte {
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	fmt.Println("GET " + url)
	if len(args) > 0 {
		url += "?" + strings.Join(args, "&")
	}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		panic(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return body
}

func HttpPost(tlsConfig *tls.Config, req *http.Request, url string, args ...string) ([]byte, error) {
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	fmt.Println("POST " + url)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func MakeArg(key string, value []byte) string {
	return key + "=" + base64.URLEncoding.EncodeToString(value)
}
