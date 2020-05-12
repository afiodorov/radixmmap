package main

func countLines(data []byte) (numLines int) {
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			numLines++
		}
	}

	if len(data) > 0 && data[len(data)-1] != '\n' {
		numLines++
	}

	return
}

func skipHeader(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	posAfterFirstNewLine := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			posAfterFirstNewLine = i + 1
			break
		}
	}

	return data[posAfterFirstNewLine:]
}
