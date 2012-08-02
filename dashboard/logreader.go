package main

import (
		"bufio"
		"io"
		"log"
		"math/rand"
		"os"
		"time"
)

func startOver(name string) (*bufio.Reader, *os.File) {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}

	return bufio.NewReader(file), file
}

func logreader(name string) {
	var line string // holds complete line, may grow

	b, f := startOver(name)
	for {
		nl, isPref, err := b.ReadLine()
		if err != nil {
			if err == io.EOF {
				f.Close()
				b, f = startOver(name)
			} else {
				log.Fatal(err)
			}
		}
		if isPref {
			line += string(nl)
			continue
		} else {
			line = string(nl)
		}
		h.broadcast <- time.Now().String() + "\t" + line
		time.Sleep(time.Duration(rand.Int63n(2500)+500)*time.Millisecond)
	}
}
