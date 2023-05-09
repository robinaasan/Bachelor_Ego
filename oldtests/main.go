package main

import (
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/robinaasan/Bachelor_Ego/test/singletest"
)

func main() {
	flag.Parse()
	args := flag.Args()
	username := args[0]
	rps := args[1]
	rps_int, err := strconv.Atoi(strings.TrimSpace(rps))
	if err != nil {
		panic("Coulndt convert number")
	}
	path := "./Go/" + rps + ".txt"

	for i := 0; i < 5; i++ {
		singletest.RunSingleTest(username, rps_int, path)
		time.Sleep(2 * time.Second)
		//os.Remove("../runtime/store.store")
	}
}
