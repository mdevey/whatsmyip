//
// "What's my ip" for alpine (musl-gcc)
// Build via:
//  CC=$(which musl-gcc) go build --ldflags '-w -linkmode external -extldflags "-static"' server.go
//

package main

import (
    "fmt"
    "net"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err == nil {
	fmt.Fprintf(w, "%s", ip)
    }
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
