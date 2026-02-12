package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("could not open messages.txt: %s\n", err)
		return
	}
	linesChan := getLinesChannel(f)
	for line := range linesChan {
		fmt.Println("read: ", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer func() {
			err := f.Close()
			fmt.Println(err)
		}()
		defer close(lines)
		currentLine := ""

		for {
			dat := make([]byte, 8)
			n, err := f.Read(dat)
			if err != nil {
				if currentLine != "" {
					lines <- currentLine
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				break
			}

			parts := strings.Split(string(dat[:n]), "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- fmt.Sprintf("%s%s", currentLine, parts[i])
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
	}()
	return lines
}
