package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
	fmt.Println("Accepted connection from", conn.RemoteAddr())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	linesChan := getLinesChannel(ctx, conn)
	for line := range linesChan {
		fmt.Println(line)
	}
	fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
}

func getLinesChannel(ctx context.Context, f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer func() {
			err := f.Close()
			if err != nil {
				fmt.Println(err)
			}
			close(lines)
		}()

		currentLine := ""

		for {
			dat := make([]byte, 8)
			n, err := f.Read(dat)
			if err != nil {
				if currentLine != "" {
					select {
					case lines <- currentLine:
					case <-ctx.Done():
					}
				}
				if !errors.Is(err, io.EOF) {
					fmt.Printf("read error: %s\n", err.Error())
				}
				return
			}

			parts := strings.Split(string(dat[:n]), "\n")
			for i := 0; i < len(parts)-1; i++ {
				fullLine := currentLine + parts[i]

				select {
				case lines <- fullLine:
					currentLine = ""
				case <-ctx.Done():
					return
				}
			}
			currentLine += parts[len(parts)-1]

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
	return lines
}
