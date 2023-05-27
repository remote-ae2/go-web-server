package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
)

const (
	WEBSOCKET_PORT = "8080"
)

func main() {
	log.Println("Launching server...")

	//config runtime
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	fmt.Printf("Running with %d CPUs\n", nuCPU)

	port := os.Getenv("PORT")
	if port == "" {
		port = WEBSOCKET_PORT
	}

	//трюк для wakemydyno.com
	//http.Handle("/", http.FileServer(http.Dir("./public")))

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		handleHttp(w, r)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebsocket(w, r)
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
