// +build cgo,!netgo

package main

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netdb.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
*/
import "C"

import (
	"flag"
	"fmt"
	"syscall"
	"unsafe"
)

func getHostByName(host string) [4]byte {
	var hints C.struct_addrinfo
	hints.ai_flags = C.AI_CANONNAME
	hints.ai_socktype = C.SOCK_STREAM
	var res *C.struct_addrinfo
	h := C.CString(host)
	defer C.free(unsafe.Pointer(h))

	// TODO: erro handling
	_, _ = C.getaddrinfo(h, nil, &hints, &res)
	defer C.freeaddrinfo(res)

	for r := res; r != nil; r = r.ai_next {
		if r.ai_socktype != C.SOCK_STREAM {
			continue
		}

		switch r.ai_family {
		case C.AF_INET:
			sa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(r.ai_addr))
			var addr [4]byte
			copy(addr[:], sa.Addr[0:4])
			return addr
		}
	}

	return [4]byte{}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		panic("illegal args")
	}
	host := args[0]

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
	defer syscall.Shutdown(fd, syscall.SHUT_RDWR)
	defer syscall.Close(fd)

	addr := getHostByName(host)
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
	header = append(header, []byte("Host: "+host+"\n")...)
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
	// TODO: split header and body
	fmt.Println(string(data))
}
