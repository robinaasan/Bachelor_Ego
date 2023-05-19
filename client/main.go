// This is a script for sending requests to the runtime
// First you need to initialise as a client to the runtime: go run main.go INIT <name>
// Secondly upload the wasm module: go run main.go UPLOAD <name>
// Finally, you can set a new key-value pair, making the runtime run the module: go run main.go SET <key> <value> <name>

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/edgelesssys/ego/attestation"
	"github.com/edgelesssys/ego/eclient"
)

const (
	usage_default string = "Usage: <cmd> <values> <clientname>"
	usage_init    string = "Usage: <clientname>"
	usage_upload  string = "Usage: <pathtofile> <clientname>"
	usage_set     string = "Usage: <key> <value> <clientname>"
)

const (
	setEndPoint    = "https://localhost:8086/Add"
	uploadEndPoint = "https://localhost:8086/Upload"
	initEndPoint   = "https://localhost:8086/Init"
)

func main() {
	fmt.Println(os.Getenv("uniqueid"))
	uniqueID, _ := hex.DecodeString("b7ce3e0e13f864d3fabd448277072b1ac5186fc96a858f747ff5eb8cbc0feda0")

	verifyReport := func(report attestation.Report) error {
		if !bytes.Equal(report.UniqueID, uniqueID) {
			return errors.New("invalid UniqueID")
		}
		return nil
	}
	tlsConfig := eclient.CreateAttestationClientTLSConfig(verifyReport) // create TLS config that verifies report from the runtime
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	err := runTerminalCommands(&client)
	if err != nil {
		fmt.Println(err)
	}
}

func runTerminalCommands(client *http.Client) error {
	flag.Parse()
	args := flag.Args()
	q := url.Values{}
	if len(args) < 1 {
		panic(usage_default)
	}
	switch args[0] {
	case "INIT":
		if len(args) < 2 { //needs "INIT" and "name"
			panic(usage_init)
		}
		q.Add("username", args[1]) // Username of the client
		err := getInitForClient(q, client)
		if err != nil {
			return err
		}
	case "UPLOAD":
		if len(args) < 3 {
			panic(usage_upload)
		}
		q.Add("username", args[2]) // name of the user
		wasmmodule := args[1]
		err := postUploadFileForClient(q, client, wasmmodule)
		if err != nil {
			return err
		}
	case "SET":
		if len(args) < 4 { //needs "SET", "key", "value", name
			panic(usage_set)
		}
		q.Add("key", args[1])
		q.Add("value", args[2])
		q.Add("username", args[3])
		err := getAdd(q, client)
		if err != nil {
			return err
		}
	default: // optimalt panic(usage)
		panic(usage_default)
	}
	return nil
}

func getAdd(q url.Values, client *http.Client) error {
	req, err := http.NewRequest("POST", setEndPoint, nil)
	if err != nil {
		return err
	}
	req.URL.RawQuery = q.Encode() // Encode and assign back to the original query.
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Res: %v\n", string(resBody))
	return nil
}

func getInitForClient(q url.Values, client *http.Client) error {
	req, err := http.NewRequest("GET", initEndPoint, nil)
	if err != nil {
		return err
	}
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Res: %v\n", string(resBody))
	return nil
}

func postUploadFileForClient(q url.Values, client *http.Client, wasmfilepath string) error {
	// https://webassembly.github.io/wabt/demo/wat2wasm/
	// 	wasmBytes := []byte(`
	//  (module
	// 		(type $t0 (func (param i32 i32) (result i32 i32 i32)))
	// 		(import "math" "set" (func $set (type $t0)))
	// 		(func $add_one (export "add_one") (type $t0) (param $x i32) (param $y i32) (result i32 i32 i32)
	// 	  		(call $set
	// 				(local.get $x)
	// 				(local.get $y))))
	// `)
	wasmBytes, err := os.ReadFile(wasmfilepath)
	if err != nil {
		return err
	}
	wasmMap := map[string][]byte{"File": wasmBytes}
	json_data, err := json.Marshal(wasmMap)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", uploadEndPoint, bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}
	req.URL.RawQuery = q.Encode() // Encode and assign back to the original query.
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// response from server:
	defer resp.Body.Close()
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Res: %v\n", string(resBody))
	return nil
}
