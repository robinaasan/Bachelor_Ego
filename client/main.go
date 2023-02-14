package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	err := uploadFile("simple.wasm", "http://localhost:8080")
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

func uploadFile(filepath string, url string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	defer file.Close()

	b := &bytes.Buffer{}

	writer := multipart.NewWriter(b)

	part, err := writer.CreateFormFile("file", filepath)

	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)

	if err != nil {
		return err
	}

	err = writer.Close()

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Byte for reading the response
	bs := make([]byte, 99999)
	resp.Body.Read(bs)
	fmt.Println(string(bs))

	defer resp.Body.Close()
	//fmt.Println(string(b.Bytes()))
	return nil
}
