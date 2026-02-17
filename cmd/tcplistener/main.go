package main

import (
	"fmt"
	"log"
	"net"

	"github.com/zzwsec/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer func() {
		err := ln.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()
	fmt.Println("Listening for TCP traffic on ", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Accepted connection from", conn.RemoteAddr())
	res, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("Error parsing request from", conn.RemoteAddr(), ":", err)
		return
	}

	fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
		res.RequestLine.Method,
		res.RequestLine.RequestTarget,
		res.RequestLine.HttpVersion)

	fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
}
