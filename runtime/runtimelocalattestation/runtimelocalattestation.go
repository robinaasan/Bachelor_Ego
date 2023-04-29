package runtimelocalattestation

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)
//code inspired from edgeless systems github report: https://github.com/edgelesssys/ego/tree/master/samples/local_attestation

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

func HttpPost(tlsConfig *tls.Config, jsonBody []byte, url string, args ...string) ([]byte, error) {
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	fmt.Println("POST " + url)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
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
