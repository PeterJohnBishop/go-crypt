package main

import (
	"fmt"
	"go-crypt/server"
)

func main() {
	fmt.Println("starting go-crypt")
	server.ServeGin()
}
