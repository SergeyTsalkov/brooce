package myip

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var ip string
var once sync.Once

func PublicIPv4() string {
	once.Do(func() {
		if IsEC2() {
			ip = ec2IP()
		} else {
			ip = externalIP()
		}

		if ip == "" {
			log.Fatalln("Unable to determine our IPv4 address!")
		}
	})

	return ip
}

func IsEC2() bool {
	data, err := ioutil.ReadFile("/sys/hypervisor/uuid")
	if err != nil {
		return false
	}

	return strings.HasPrefix(string(data), "ec2")
}

func ec2IP() string {
	// this is ugly, but I don't think there's a better way to get our IP on EC2
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Get("http://169.254.169.254/latest/meta-data/public-ipv4")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	buf := bytes.Buffer{}
	buf.ReadFrom(resp.Body)

	return strings.TrimSpace(buf.String())
}

func externalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return ""
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String()
		}
	}
	return ""
}
