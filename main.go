package main

import "net/http"

func main() {
	startServer()
}

func startServer() {
	var server http.Server
	serverHandle := http.NewServeMux()
	serverHandle.Handle("/", http.FileServer(http.Dir(".")))
	server.Handler = serverHandle

	server.Addr = "localhost:8080"

	server.ListenAndServe()

}
