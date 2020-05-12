package main

import (
	"log"
	"os"

	mm "github.com/edsrzf/mmap-go"
	"github.com/jfcg/sixb"
)

//FileMeta contains data of the file and number of lines needed to allocate extra memory necessary for sorting
type FileMeta struct {
	Content  []byte
	NumLines int
}

func memoryMapFile(src *os.File, res []FileMeta, index int, shouldSkipHeader bool) {
	data, err := mm.Map(src, mm.RDONLY, 0)
	if shouldSkipHeader {
		data = skipHeader(data)
	}

	if err != nil {
		log.Fatalf("couldn't mmap file: %v\n", err)
	}

	numLines := countLines(data)

	res[index] = FileMeta{
		Content:  data,
		NumLines: numLines,
	}
}

func lineSplitter(lines Lines, fileContent []byte, numLinesBefore int) {
	lineStart := 0
	lineNumber := numLinesBefore
	for i := 0; i < len(fileContent); i++ {
		if fileContent[i] == '\n' {
			lines[lineNumber] = sixb.BtS(fileContent[lineStart : i+1])
			lineStart = i + 1
			lineNumber++
		}
	}

	if len(fileContent) > lineStart {
		lines[lineNumber] = sixb.BtS(fileContent[lineStart:]) + "\n"
	}
}
