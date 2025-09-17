package main

import (
	"flag"
	"log"

	"github.com/CyberOakAlpha/CrossNet/internal/web"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "Port to run the web server on")
	flag.IntVar(&port, "p", 8080, "Port to run the web server on (short)")
	flag.Parse()

	server := web.NewServer(port)
	log.Fatal(server.Start())
}