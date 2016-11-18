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

func ipHandler(w http.ResponseWriter, r *http.Request) {
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err == nil {
        fmt.Fprintf(w, "%s", ip)
    }
}


func dnsHandler(w http.ResponseWriter, r *http.Request) {
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err == nil {
        //TODO aliases
        //TODO fqdn instead of 'localhost' - client & server on same host
        names, err := net.LookupAddr(ip);
        if err == nil {
            for _,v := range names {
                //truncate trailing '.' that may be appended.
                last := len(v)-1
                out := v;
                if last >= 0 && v[last] == '.' {
                    out = v[:last]
                }

                fmt.Fprintf(w, "%s\n", out)
            }
        }
    }
}

func main() {
    http.HandleFunc("/", ipHandler)
    http.HandleFunc("/dns", dnsHandler)
    http.ListenAndServe(":8080", nil)
}
