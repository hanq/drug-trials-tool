package main

import (
    "flag"
    "log"
    "net/http"
)

func main() {
    var port int
    flag.IntVar(&port, "port", 8080, "HTTP server port")
    flag.Parse()

    mux := http.NewServeMux()
    mux.HandleFunc("/api/spa/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte({"status":"ok"}))
    })

    addr := fmt.Sprintf(":%d", port)
    log.Printf("SPA API server starting on %s", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
}
