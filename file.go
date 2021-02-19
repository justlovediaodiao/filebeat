package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"
)

// FileBeater filebeat.
type FileBeater struct {
	ID       string // file id
	Path     string
	Offset   int64
	interval time.Duration
	file     *os.File
	scanner  *Scanner
	done     bool
}

// FileUID return file unique id.
func FileUID(info os.FileInfo) string {
	st := info.Sys().(*syscall.Stat_t)
	return fmt.Sprintf("%d/%d", st.Dev, st.Ino)
}

// NewFileBeater return a FileBeater.
func NewFileBeater(id string, path string, offset int64, interval time.Duration) (*FileBeater, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if offset > 0 {
		if _, err = file.Seek(offset, 0); err != nil {
			return nil, err
		}
	}
	beater := &FileBeater{
		ID:       id,
		Path:     path,
		Offset:   offset,
		interval: interval,
		file:     file,
		scanner:  NewScanner(file),
	}
	return beater, nil
}

// ReadLine read a line of file, ignore empty line. If not new data, wait until read a line.
// It returns io.EOF if haverst ended.
func (f *FileBeater) ReadLine() ([]byte, error) {
	eof := 0
	for {
		buf, err := f.scanner.Scan()
		if err != nil {
			if err == ErrScannerStopped {
				return nil, io.EOF
			}
			return nil, err
		}
		if buf != nil {
			eof = 0
			f.Offset += int64(len(buf))
			buf = dropCRLF(buf)
			if len(buf) == 0 {
				continue
			}
			return buf, nil
		}
		// EOF
		eof++
		if eof >= 100 { // too many EOF, check file status.
			eof = 0
			switch f.stat() {
			case -1: // harvest ended, stop scan, read remain buf.
				log.Printf("%s moved or deleted, scan stop", f.Path)
				f.scanner.Stop()
				continue
			case 0: // seek to file head and rescan.
				log.Printf("%s truncated, seek to head", f.Path)
				if _, err = f.file.Seek(0, 0); err != nil {
					return nil, err
				}
				f.Offset = 0
				f.scanner = NewScanner(f.file)
			}
		}
		time.Sleep(f.interval)
	}
}

// Close close file.
func (f FileBeater) Close() error {
	return f.file.Close()
}

// stat report harvest status.
// -1: finished
// 0: truncated
// 1: continue
func (f FileBeater) stat() int {
	info, err := os.Stat(f.Path)
	if err != nil {
		if os.IsNotExist(err) { // deleted
			return -1
		}
		return 1
	}
	if f.ID != FileUID(info) { // moved
		return -1
	}
	if f.Offset > info.Size() { // truncated
		return 0
	}
	return 1
}
