package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	p := flag.String("port", "8080", "port to run the file server on")
	flag.Parse()
	port := *p
	log.Printf("Fileserver started at port %s", port)
	log.Printf("Open your browser at http://localhost:%s", port)
	http.ListenAndServe(":"+port, http.FileServer(http.Dir(".")))
}
