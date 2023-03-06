package main

import (
	"fmt"
	"os"
)

func main() {
	//os.Setenv("GEEKS", "geeeek!!!!!!!!!!!!!")
	fmt.Println(os.Environ())
}
