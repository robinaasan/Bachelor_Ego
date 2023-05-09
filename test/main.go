package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edgelesssys/ego/attestation"
	"github.com/edgelesssys/ego/eclient"
)

type Transaction struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	ClientName string `json:"ClientName"`
}

type SetValue struct {
	Key    int
	NewVal int
}

type ResponsesRuntime struct {
	response string
	endpoint string
	err      error
	vals     SetValue
	duration time.Duration
}

// var endpoints = []string{"http://localhost:8087", "http://localhost:8087", "http://localhost:8087"}

func main() {
	// setval := &SetValue{2, 1}

	uniqueID, _ := hex.DecodeString("8ef8ee741751bc4714b548fb69ab8bccbe95b1f3387800708032941adddfc728")

	verifyReport := func(report attestation.Report) error {
		if !bytes.Equal(report.UniqueID, uniqueID) {
			return errors.New("invalid UniqueID")
		}
		return nil
	}
	tlsConfig := eclient.CreateAttestationClientTLSConfig(verifyReport)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	//client := &http.Client{}

	var c chan time.Duration
	// c = []time.Duration{}
	c = make(chan time.Duration)

	//timerChan := make(chan bool)
	//wg.Add(1)

	const orderingURL = "https://localhost:8090/Add"
	//var storeResponse []string

	flag.Parse()
	args := flag.Args()
	username := args[0]
	//username1 := args[1]
	//username2 := args[2]

	//userList := []string{username, username1, username2}

	//timer := time.NewTimer(1 * time.Second)
	var key, value int
	key = 1
	value = 1

	routines := 2

	count := 0

	// for i := 0; i < 2; i++ {

	wg := &sync.WaitGroup{}
	//var wl sync.Mutex
	//timer := time.NewTimer(1 * time.Second)
	wg.Add(routines)
	go func() {
		for i := 0; i < routines; i++ {
			go sendConcReq(key, value, orderingURL, client, c, username, time.NewTimer(1*time.Second), wg, i)
		}
	}()

	//nWg := &sync.WaitGroup{}
	listRespTime := []string{}

	// go func(){
	// 	wg.Wait()
	// 	close(c)
	// }()

	//nWg.Add(1)
	go func() {
		for t := range c {
			count++
			fmt.Println(t)
			listRespTime = append(listRespTime, strings.Replace(t.String(), "ms", "", 1))
			fmt.Println(count)
		}
	}()
	//nWg.Done()
	//nWg.Wait()
	wg.Wait()

	time.Sleep(1 * time.Second)
	storeDataInFile(&listRespTime)

	close(c)
	fmt.Println(count)

	//close(c)

}

func sendConcReq(key int, value int, orderingURL string, client *http.Client, c chan time.Duration, username string, timer *time.Timer, wg *sync.WaitGroup, i int) {
	// for {
	// 	//fmt.Printf(".")
	// 	select {
	// 	case <-timer.C:
	//
	// 		return
	// 	default:
	// 		sendToRuntime(key, value, orderingURL, client, c, time.Now(), username, i)
	// 	}
	// }
	for i := 0; i < 10000; i++ {
		sendToRuntime(key, value, orderingURL, client, c, time.Now(), username, i)
	}
	wg.Done()
}

func storeDataInFile(data *[]string) error {
	f, err := os.OpenFile("./storeResponseInFile.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o777)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for _, times := range *data {
		if _, err := f.WriteString(times + ", "); err != nil {
			panic(err)
		}
	}

	// timeDiff := endTime.Sub(time.UnixMicro(blockFromTransactions.TimeStamp)).Microseconds()

	// os.Remove("storeResponseInFile.txt")
	// err := os.WriteFile("storeResponseInFile.txt", []byte(toString(data)), 0o777)
	// if err != nil {
	// 	return err
	// }
	fmt.Println("Success writing to the file!")
	return nil
}

func toString(data *[]string) string {
	return strings.Join([]string(*data), ",")
}

// func decrementCounter(wg *sync.WaitGroup) {
// 	defer (*wg).Done()
// }

func sendToRuntime(key int, value int, endpoint string, runtime *http.Client, c chan time.Duration, tm time.Time, username string, i int) {
	//defer wg.Done()
	// fmt.Println(value)

	// t := Transaction{
	// 	ClientName: "robin",
	// 	Key:        key,
	// 	NewVal:     value,
	// }
	// q := url.Values{}
	// body := map[string]int{"Key": setvalues.Key, "NewVal": setvalues.NewVal, "OldVal": setvalues.OldVal}
	// q.Add("client", nameClient)
	// jsonBody, err := json.Marshal(t)
	// if err != nil {
	// 	c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
	// 	return
	// }
	q := &url.Values{}
	q.Add("username", username)
	q.Add("key", fmt.Sprint(i))
	q.Add("value", fmt.Sprint(value))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		//c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
		fmt.Println(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = q.Encode()
	res, err := runtime.Do(req)
	if err != nil {
		// 	c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
		fmt.Println(res.StatusCode)
		return
	}
	defer res.Body.Close()
	//responseData, err := io.ReadAll(res.Body)
	if err != nil {
		//c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
		panic("Error posting")
	}
	//fmt.Println(string(responseData))
	var duration time.Duration = time.Since(tm)
	//fmt.Println(duration)
	//c <- ResponsesRuntime{endpoint, res.Status, nil, SetValue{Key: key, NewVal: value}, duration}
	//*c = append(*c, ResponsesRuntime{endpoint, res.Status, nil, SetValue{Key: key, NewVal: value}, duration})
	c <- duration
}
