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
	usage_set    string = "Usage: client <cmd> <key> <value>"
	usage_upload string = "Usage: client <UPLOAD> <USERNAME>"
)

const (
	addEndPoint    = "https://localhost:8086/Add"
	uploadEndPoint = "https://localhost:8086/Upload"
	initEndPoint   = "https://localhost:8086/Init"
)

// type Client struct {
// 	c    *http.Client
// 	name string
// }

// func NewClient() *Client {
// 	c := &http.Client{}
// 	return &Client{
// 		c:    c,
// 		name: "Robs",
// 	}
// }
// var userHash []byte

func main() {
	uniqueID, _ := hex.DecodeString("94541bd9bc3570e4e7486d7dbea6a9f6f9c535f3d938ec776a7eb89084f43a84")

	verifyReport := func(report attestation.Report) error {
		if !bytes.Equal(report.UniqueID, uniqueID) {
			return errors.New("invalid UniqueID")
		}
		return nil
	}
	tlsConfig := eclient.CreateAttestationClientTLSConfig(verifyReport)
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	//client := &http.Client{}
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
		panic(usage_set)
	}
	switch args[0] {

	case "INIT":
		// TODO: usage
		if len(args) < 2 {
			panic(usage_set)
		}
		q.Add("username", args[1]) // Username of the client
		err := getAndStoreHash(q, client)
		if err != nil {
			return err
		}
	case "SET":
		if len(args) < 4 {
			panic(usage_set)
		}
		q.Add("cmd", "SET")
		q.Add("key", args[1])
		q.Add("value", args[2])
		q.Add("username", args[3])
		err := getAdd(q, client)
		if err != nil {
			return err
		}
		// run function that calls one endpoint

	case "UPLOAD":
		if len(args) < 2 {
			panic(usage_upload)
		}
		q.Add("cmd", "UPLOAD")
		q.Add("username", args[1]) // name of the user
		err := postUploadFile(q, client, args[1])
		if err != nil {
			return err
		}
		// run function that calls the other one
	default: // optimalt panic(usage)
		panic(usage_set)
	}
	return nil
}

func getAdd(q url.Values, client *http.Client) error {
	req, err := http.NewRequest("POST", addEndPoint, nil)
	if err != nil {
		return err
	}
	req.URL.RawQuery = q.Encode() // Encode and assign back to the original query.

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// response from server:
	// bs := make([]byte, 1024)
	// resp.Body.Read(bs)
	// fmt.Printf("%v\n", string(bs))
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Res: %v\n", string(resBody))

	return nil
}

func getAndStoreHash(q url.Values, client *http.Client) error {
	// b := []byte{}
	req, err := http.NewRequest("GET", initEndPoint, nil)
	if err != nil {
		return err
	}
	req.URL.RawQuery = q.Encode()
	//(*client).Timeout = 5 * time.Second
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

func postUploadFile(q url.Values, client *http.Client, name string) error {
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
	wasmBytes, err := os.ReadFile("./newwasm.wasm")
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
