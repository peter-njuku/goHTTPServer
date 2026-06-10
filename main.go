package main

import (
	//"image"
	"log"
	"net/http"
)

func main() {
	port := "8080"
	serveMux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	serveMux.Handle("/pages", http.FileServer(http.Dir(".")))

	serveMux.Handle("/", http.FileServer(http.Dir(".")))

	log.Printf("Serving files from root on port %s\n", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Print(err)
	}
}
