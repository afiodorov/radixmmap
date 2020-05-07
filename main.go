package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	mm "github.com/edsrzf/mmap-go"
	"github.com/twotwotwo/sorts"
)

type Lines []mm.MMap

func (s Lines) Less(i, j int) bool {
	limI, limJ := 20, 20

	if len(s[i]) < limI {
		limI = len(s[i])
	}

	if len(s[j]) < limJ {
		limJ = len(s[j])
	}

	return bytes.Compare(s[i][:limI], s[j][:limJ]) == -1
}

func (s Lines) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Lines) Len() int {
	return len(s)
}

func (s Lines) Key(i int) []byte {
	l := 20
	if len(s[i]) < l {
		l = len(s[i])
	}

	return s[i][:l]
}

func (s Lines) Sort() { sorts.ByBytes(s) }

func main() {
	fileName := flag.String("s", "", "file to sort")
	flag.Parse()

	file, err := os.Open(*fileName)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("%v\n", err)
		}
	}()

	m, err := mm.Map(file, mm.RDONLY, 0)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	defer func() {
		if err := m.Flush(); err != nil {
			log.Fatalf("%v\n", err)
		}
	}()

	var lines Lines

	start := 0

	for i := 0; i < len(m); i++ {
		if m[i] == byte(10) {
			lines = append(lines, m[start:i])
			start = i + 1
		}
	}

	if len(m) > start {
		lines = append(lines, m[start:len(m)])
	}

	lines.Sort()

	for _, l := range lines {
		fmt.Printf("%s\n", l)
	}
}
