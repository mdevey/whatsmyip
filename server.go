//
// "What's my ip" for alpine (musl-gcc)
// Build via:
//  CC=$(which musl-gcc) go build --ldflags '-w -linkmode external -extldflags "-static"' server.go
//

package main

import (
    "flag"
    "fmt"
    "net"
    "net/http"
    "log"
    "strings"
)

var myip string

func ipHandler(w http.ResponseWriter, r *http.Request) {
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err == nil {
        if ip == "127.0.0.1" ||
           ip == "::1" ||
           isPrivateSubnet(net.ParseIP(ip)) { //local docker containers etc
           fmt.Fprintf(w, "%s", myip)
        } else {
           fmt.Fprintf(w, "%s", ip)
        }
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

func getIPNet(network string) *net.IPNet {
    _, subnet, _ := net.ParseCIDR(network)
    return subnet;
}

var privateCIDRs = []*net.IPNet {
    getIPNet("10.0.0.0/8"),
    getIPNet("172.16.0.0/12"),
    getIPNet("192.168.0.0/16"),
}

//If 'myip' is in a private range, we assume you are using this service to find ip's in that range also.
//TODO allow _multiple_ private ranges (most companies just use one, typically 10.0.0.0 or 192.168.0.0)
func AllowPrivateSubnetForMyip() *net.IPNet {
    var ip= net.ParseIP(myip)
    for i, ipr := range privateCIDRs {
      if ipr.Contains(ip) {
          //remove ipr
          privateCIDRs = append(privateCIDRs[:i], privateCIDRs[i+1:]...)
          return ipr
      }
    }
    return nil
}

// isPrivateSubnet - check to see if this ip is in a private subnet
func isPrivateSubnet(ip net.IP) bool {
    for _, r := range privateCIDRs {
      if r.Contains(ip) {
          return true
      }
    }
    return false
}

// Get preferred outbound ip of this machine
func getOutboundIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().String()
    idx := strings.LastIndex(localAddr, ":")

    return localAddr[0:idx]
}

func main() {
    flag.Parse()

    if flag.NArg() < 1 {
        myip = getOutboundIP()
        log.Println("Warning: Blindly choosing my IP of " + myip)
    } else {
        myip = flag.Arg(0)
    }
    log.Println("For IP's in private subnets I will return my IP: " + myip)

    network := AllowPrivateSubnetForMyip()
    if network != nil {
      log.Println(network.String() + " private IPs are the exception and will also be returned.")
    }

    http.HandleFunc("/", ipHandler)
    http.HandleFunc("/dns", dnsHandler)
    http.ListenAndServe(":8080", nil)
}
