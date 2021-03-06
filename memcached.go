package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	i := Interpreter{r: os.Stdin, w: os.Stdout}
	i.Loop()
}

type Interpreter struct {
	r io.Reader
	w io.Writer
}

func (i *Interpreter) Loop() {
	s := bufio.NewScanner(i.r)
	s.Split(ScanTextLines)
	for s.Scan() {
		line := s.Text()
		fmt.Printf("text (%d): %s\n", len(line), line)
		cmd, err := Parse(line)
		if err != nil {
			fmt.Printf("parse error: %v\n", err)
		} else {
			fmt.Printf("command: %v\n", cmd)
		}
		n, err := io.WriteString(i.w, "ok\r\n")
		if err != nil {
			fmt.Printf("write error: %v\n", err)
		} else {
			fmt.Printf("%d bytes written\n", n)
		}
	}
	if err := s.Err(); err != nil {
		fmt.Printf("scan error: %v\n", err)
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

type Command interface {
	isCommand() bool
}

type Set struct {
	key     string
	flags   uint16
	exptime time.Time
	bytes   uint64
	noreply bool
}

type Get struct {
	keys []string
}

func (s Set) isCommand() bool {
	return true
}

func (s Set) String() string {
	return fmt.Sprintf("set %s %d %v %d %t", s.key, s.flags, s.exptime, s.bytes, s.noreply)
}

func (g Get) String() string {
	return fmt.Sprintf("get %v", g.keys)
}

func (g Get) isCommand() bool {
	return true
}

func Parse(s string) (Command, error) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil, errors.New("empty fields")
	}
	switch name := fields[0]; name {
	case "get":
		keys := fields[1:]
		if len(keys) == 0 {
			return nil, errors.New("get: no keys")
		}
		g := Get{keys: keys}
		return &g, nil
	case "set":
		params := fields[1:]
		if len(params) < 4 || len(params) > 5 {
			return nil, errors.New("set: invalid parameter count")
		}
		s := Set{}
		s.key = params[0]
		flags, err := strconv.ParseUint(params[1], 10, 16)
		if err != nil {
			return nil, errors.New("set: invalid flags field")
		}
		s.flags = uint16(flags)
		exptime, err := strconv.ParseUint(params[2], 10, 32)
		if err != nil {
			return nil, errors.New("set: invalid exptime field")
		}
		if exptime <= 2592000 { // number of seconds in 30 days
			s.exptime = time.Now().Add(time.Duration(exptime) * time.Second)
		} else {
			s.exptime = time.Unix(int64(exptime), 0)
		}
		bytes, err := strconv.ParseUint(params[3], 10, 64)
		if err != nil {
			return nil, errors.New("set: invalid bytes (size) field")
		}
		s.bytes = bytes
		if len(params) == 5 {
			if params[4] == "noreply" {
				s.noreply = true
			} else {
				return nil, errors.New("set: invalid noreply field")
			}
		} else {
			s.noreply = false
		}
		return &s, nil
	}
	return nil, errors.New("unrecognized command")
}
