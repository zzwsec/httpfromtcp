package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	dat := make([]byte, 8)
	for {
		n, err := f.Read(dat)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("read: %s\n", string(dat[:n]))
	}
}
