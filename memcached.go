package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
)

func main() {
	r := strings.NewReader("hello\r\nworld\r\n")
	s := bufio.NewScanner(r)
	s.Split(ScanTextLines)
	for s.Scan() {
		line := s.Text()
		fmt.Printf("text (%d): %s\n", len(line), line)
	}
	if err := s.Err(); err != nil {
		fmt.Printf("scan error: %s\n", err)
	}
}

var Delimiter = []byte{'\r', '\n'}

// ScanLines is a split function for a Scanner that returns each line of
// text terminated by a trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is one carriage return followed
// by one newline. In regular expression notation, it is '\r\n'.
func ScanTextLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, Delimiter); i >= 0 {
		// We have a full terminated line.
		return i + 2, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return
	// an error.
	if atEOF {
		return 0, nil, errors.New("unexpected EOF")
	}
	// Request more data.
	return 0, nil, nil
}
