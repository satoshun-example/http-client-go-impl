package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
	"syscall"
)

var host = flag.String("host", "", "hostname")

func main() {
	flag.Parse()

	if *host == "" {
		panic("no specify hostname")
	}

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(fd)

	// TODO: remove net package
	addrs, err := net.LookupHost(*host)
	if err != nil {
		panic(err)
	}

	var addr [4]byte
	t := addrs[0]
	tt := strings.Split(t, ".")
	for i, v := range tt {
		vv, _ := strconv.Atoi(v)
		addr[i] = byte(vv)
	}

	inet4 := &syscall.SockaddrInet4{
		Port: 80,
		Addr: addr,
	}

	err = syscall.Connect(fd, inet4)
	if err != nil {
		panic(err)
	}

	syscall.Write(fd, []byte("GET / HTTP/1.1\n"))
	syscall.Write(fd, []byte("Host: "+*host+"\n"))
	syscall.Write(fd, []byte("User-Agent: Test Client\n"))
	syscall.Write(fd, []byte("Accept: */*\n\n"))

	data := make([]byte, 0)
	for {
		d := make([]byte, 512)
		n, err := syscall.Read(fd, d)
		if err != nil {
			panic(err)
		}
		if n == 0 {
			break
		}
		data = append(data, d[:n]...)
	}
	log.Println(string(data))
}
