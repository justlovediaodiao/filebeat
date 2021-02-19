// Thanks to The Golang Standard Library bufio/scan.go
package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

const (
	maxScanTokenSize = 64 * 1024 // Maximum size used to buffer.
	startBufSize     = 4096      // Size of initial allocation for buffer.
)

// ErrScannerStopped returned when stop scanner manually.
var ErrScannerStopped = errors.New("scan stopped")

// Scanner is similar to bufio.Scanner. But it returns nil,nil when no new line or EOF.
// It can be used on a consistent new data arriving reader.
type Scanner struct {
	r     io.Reader // The reader provided by the client.
	buf   []byte    // Buffer used as argument to split.
	start int       // First non-processed byte in buf.
	end   int       // End of data in buf.
	err   error     // Sticky error.
}

// NewScanner returns a new Scanner to read from r.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r:   r,
		buf: make([]byte, startBufSize),
	}
}

// Stop stop the scanner. Scanner will not read any more.
// After stopped, You should call Scan until it returns an error to get remain buf data.
// The returned error of Scan call is ErrScannerStopped, or any other error (except EOF) happend before Stop call.
func (s *Scanner) Stop() {
	if s.err == nil {
		s.err = ErrScannerStopped
	}
}

// Scan scan for a line. Returned lines are not trimmed by '\n'.
// If no new line or EOF error, return nil,nil.
// You can call Scan later for new arrived data, or Call Stop to finish scan.
// If returned error is not nil, []byte is nil and scan finished.
func (s *Scanner) Scan() ([]byte, error) {
	for {
		// find '\n' from buf.
		if s.end > s.start {
			if i := bytes.IndexByte(s.buf[s.start:s.end], '\n'); i != -1 {
				i++
				data := s.buf[s.start : s.start+i]
				s.start += i
				return data, nil
			}
		}
		// not found and error occurred, should not read any more. return buf data and then return error.
		if s.err != nil {
			if s.end > s.start {
				data := s.buf[s.start:s.end]
				s.start = s.end
				return data, nil
			}
			return nil, s.err
		}
		// move buf data to buf head, so we can read more data one time.
		if s.start > 0 && (s.end == len(s.buf) || s.start > len(s.buf)/2) {
			copy(s.buf, s.buf[s.start:s.end])
			s.end -= s.start
			s.start = 0
		} else if s.end == len(s.buf) { // s.start == 0. buf full, resize.
			if len(s.buf) == maxScanTokenSize { // line too long, overflow buf size. Is there any better way rather than return an error?
				return s.buf, bufio.ErrBufferFull
			}
			newSize := len(s.buf) * 2
			if newSize > maxScanTokenSize {
				newSize = maxScanTokenSize
			}
			newBuf := make([]byte, newSize)
			copy(newBuf, s.buf)
			s.buf = newBuf
		}
		// read more data.
		n, err := s.r.Read(s.buf[s.end:len(s.buf)])
		s.end += n
		if err != nil && err != io.EOF { // we got error, continue to Scan to finish.
			s.err = err
			continue
		}
		if n == 0 { // read nothing. This means you should call Scan later to try reading new data.
			return nil, nil
		}
	}
}

// dropCRLF drop '\r' or '\n' or '\r\n' at the end of line.
func dropCRLF(data []byte) []byte {
	n := len(data) - 1
	if n >= 0 && data[n] == '\n' {
		n--
	}
	if n >= 0 && data[n] == '\r' {
		n--
	}
	return data[:n+1]
}
