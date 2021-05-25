package main

import (
	"flag"
	"fmt"
	"log"
)
func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	fmt.Println("test")
	port := flag.Uint("port", 5000, "TCP Port Number for Blockchain Server")
	flag.Parse()
	fmt.Println(*port)
	app := NewBlockchainServer(uint16(*port))
	app.Run()
}