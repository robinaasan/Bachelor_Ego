package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
)

// MyStruct is an example structure for this program.
type MyStruct struct {
    Key string `json:"Key"`
    Val string `json:"Val"`
}

func main() {
    http.HandleFunc("/", handlerTransaction)

    server := http.Server{Addr: "localhost:8082"}
    fmt.Println("Listening...")
    err := server.ListenAndServe()
    fmt.Println(err)

}

func handlerTransaction(w http.ResponseWriter, r *http.Request) {
    // noe
    // Check if file exist, else create one

    filename := "testServe.json"
    err := checkFile(filename)
    if err != nil {
        panic(err)
    }

    updateBlockchain(filename, "hei", "iver")

    fmt.Println("Funker med client")
}

func checkFile(filename string) error {
    _, err := os.Stat(filename)
    if os.IsNotExist(err) {
        _, err := os.Create(filename)
        if err != nil {
            return err
        }
    }
    return nil
}

func updateBlockchain(filename string, key string, val string) {

    file, err := os.ReadFile(filename)
    if err != nil {
        panic(err)
    }

    data := []MyStruct{}

    // Here the magic happens!
    json.Unmarshal(file, &data)

    newStruct := &MyStruct{
        Key: key,
        Val: val,
    }

    data = append(data, *newStruct)
    fmt.Printf("%+v", data)

    // Preparing the data to be marshalled and written.
    dataBytes, err := json.Marshal(data)
    if err != nil {
        panic(err)
    }

    err = ioutil.WriteFile(filename, dataBytes, 0644)
    if err != nil {
        panic(err)
    }
}
