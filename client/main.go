package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
)

const usage_add string = "Usage: client <cmd> <key> <value>"
const usage_upload string = "Usage: client <upload> <file>"

const addEndPoint = "http://localhost:8080/Add"
const uploadEndPoint = "http://localhost:8080/Upload"

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
	// resp, err := http.Get("http://localhost:8080/monkeys")
	// if err != nil {
	// 	panic("wups")
	// }
	// bs := make([]byte, 99999)
	// resp.Body.Read(bs)
	// fmt.Println(string(bs))
}

func runTerminalCommands(client *Client) error {

	//serverURL := flag.String("url", "localhost:8080", "Server's url")
	flag.Parse()
	args := flag.Args()
	//req := url.URL{Scheme: "http", Host: *serverURL, Path: "/secret"}

	q := url.Values{}

	if len(args) < 1 {
		panic(usage_add)
		// } else { // bare for å sjekke
		// 	fmt.Printf(args[0])
		// 	fmt.Printf(args[1])
	}

	switch args[0] {
	case "add":
		if len(args) < 3 {
			panic(usage_add)
		}
		q.Add("cmd", "add")
		q.Add("val1", args[1])
		q.Add("val2", args[2])

		err := getAdd(q, client)

		if err != nil {
			return err
		}
		//run function that calls one endpoint

	case "upload":
		if len(args) < 1 {
			panic(usage_upload)
		}
		q.Add("cmd", "upload")
		//q.Add("filename", args[1])

		err := postUploadFile(q, client)
		if err != nil {
			return err
		}

		//run function that calls the other one
	default: // optimalt panic(usage)
		q.Add("cmd", "noe")
		q.Add("val1", args[1])
		q.Add("val2", args[2])

	}

	//fmt.Println(q)

	return nil
	// Byte for reading the response

}

func getAdd(q url.Values, client *Client) error {
	b := &bytes.Buffer{}

	req, err := http.NewRequest("POST", addEndPoint, b)
	//req.RawQuery = q.Encode()
	req.URL.RawQuery = q.Encode() // Encode and assign back to the original query.

	//client := &http.Client{}
	resp, err := client.c.Do(req)
	if err != nil {
		return err
	}

	//response from server:
	bs := make([]byte, 99999)
	resp.Body.Read(bs)
	fmt.Printf("%v\n", string(bs))

	defer resp.Body.Close()
	return nil
}

func postUploadFile(q url.Values, client *Client) error {
	wasmBytes := []byte(`
	(module
		;; We import a math.sum function.
		(import "math" "sum" (func $sum (param i32 i32) (result i32)))

		;; We export an add_one function.
		(func (export "add_one") (param $x i32) (result i32)
			local.get $x
			i32.const 1
			call $sum))
`)
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
	//req.RawQuery = q.Encode()
	if err != nil {
		return err
	}
	req.URL.RawQuery = q.Encode() // Encode and assign back to the original query.
	req.Header.Set("Content-Type", "application/json")

	//client := &http.Client{}
	resp, err := client.c.Do(req)
	if err != nil {
		return err
	}

	//response from server:
	bs := make([]byte, 99999)
	resp.Body.Read(bs)
	fmt.Printf("%v\n", string(bs))

	defer resp.Body.Close()
	//fmt.Println(string(b.Bytes()))
	return nil
}
