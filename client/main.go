package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Running client on port :8080...")
	err := http.ListenAndServe(":8080", http.FileServer(http.Dir("data")))
	if err != nil {
		panic(err)
	}
}
