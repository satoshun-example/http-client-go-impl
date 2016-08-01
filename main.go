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
	defer syscall.Shutdown(fd, syscall.SHUT_RDWR)
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

	// write request heaader
	var header []byte
	header = append(header, []byte("GET / HTTP/1.1\n")...)
	header = append(header, []byte("Host: "+*host+"\n")...)
	header = append(header, []byte("User-Agent: Test Client\n")...)
	header = append(header, []byte("Accept: */*\n\n")...)
	syscall.Write(fd, header)

	syscall.Shutdown(fd, syscall.SHUT_WR)

	var data []byte
	for {
		d := make([]byte, 255)
		n, err := syscall.Read(fd, d)
		if n == 0 {
			break
		}
		if err != nil {
			panic(err)
		}
		data = append(data, d[:n]...)
	}
	log.Println(string(data))
}
