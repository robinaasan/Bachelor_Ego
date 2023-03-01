package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const usage_set string = "Usage: client <cmd> <key> <value>"
const usage_upload string = "Usage: client <upload> <file>"

const addEndPoint = "http://localhost:8081/Add"
const uploadEndPoint = "http://localhost:8081/Upload"

type Client struct {
	c    *http.Client
	name string
}

func NewClient() *Client {
	c := &http.Client{}
	return &Client{
		c:    c,
		name: "Robs",
	}
}

func main() {

	client := NewClient()

	err := runTerminalCommands(client)

	if err != nil {
		fmt.Println(err)
	}
}

func runTerminalCommands(client *Client) error {

	flag.Parse()
	args := flag.Args()
	q := url.Values{}

	if len(args) < 1 {
		panic(usage_set)
	}

	switch args[0] {
	case "SET":
		if len(args) < 3 {
			panic(usage_set)
		}
		q.Add("cmd", "SET")
		q.Add("key", args[1])
		q.Add("value", args[2])

		err := getAdd(q, client)

		if err != nil {
			return err
		}
		//run function that calls one endpoint

	case "UPLOAD":
		if len(args) < 1 {
			panic(usage_upload)
		}
		q.Add("cmd", "UPLOAD")
		err := postUploadFile(q, client)
		if err != nil {
			return err
		}

		//run function that calls the other one
	default: // optimalt panic(usage)
		panic(usage_set)
	}

	return nil
}

func getAdd(q url.Values, client *Client) error {
	b := &bytes.Buffer{}

	req, err := http.NewRequest("POST", addEndPoint, b)
	req.URL.RawQuery = q.Encode() // Encode and assign back to the original query.

	resp, err := client.c.Do(req)
	if err != nil {
		return err
	}

	//response from server:
	bs := make([]byte, 1024)
	resp.Body.Read(bs)
	fmt.Printf("%v\n", string(bs))

	defer resp.Body.Close()
	return nil
}

func postUploadFile(q url.Values, client *Client) error {
	// https://webassembly.github.io/wabt/demo/wat2wasm/
	// 	wasmBytes := []byte(`
	// 	(module
	// 		;; We import a math.set function.
	// 		(import "math" "set" (func $set (param i32 i32) (result i32)))

	// 		;; We export an add_one function.
	// 		(func (export "add_one") (param $x i32) (param $y i32) (result i32)
	// 			local.get $x
	// 			local.get $y
	// 			call $set))
	// `)
	wasmBytes, err := os.ReadFile("./test.wasm")
	if err != nil {
		return err
	}
	wasmMap := map[string][]byte{"File": wasmBytes}
	json_data, err := json.Marshal(wasmMap)

	//file, err := os.Open(filepath)
	// if err != nil {
	// 	return err
	// }

	//defer file.Close()

	//b := &bytes.Buffer{}

	//writer := multipart.NewWriter(b)

	//part, err := writer.CreateFormFile("file", filepath)

	// if err != nil {
	// 	return err
	// }

	// _, err = io.Copy(part, file)

	// if err != nil {
	// 	return err
	// }

	// err = writer.Close()

	// if err != nil {
	// 	return err
	// }

	// if err != nil {
	// 	return err
	// }
	req, err := http.NewRequest("POST", uploadEndPoint, bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}
	req.URL.RawQuery = q.Encode() // Encode and assign back to the original query.
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.c.Do(req)
	if err != nil {
		return err
	}

	//response from server:
	bs := make([]byte, 512)
	resp.Body.Read(bs)
	fmt.Printf("%v\n", string(bs))

	defer resp.Body.Close()
	//fmt.Println(string(b.Bytes()))
	return nil
}
