/*
Author: Yasuhiro Matsumoto (a.k.a mattn)
Origin: https://github.com/mattn/goreman
Licence: MIT
Changes: package name, exclude go-colorable dependencies
Changed by Vlad Krinitsyn
*/
package vman

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"
)

type clogger struct {
	idx     int
	name    string
	writes  chan []byte
	done    chan struct{}
	timeout time.Duration // how long to wait before printing partial lines
	buffers buffers       // partial lines awaiting printing
}

var colors = []int{
	32, // green
	36, // cyan
	35, // magenta
	33, // yellow
	34, // blue
	31, // red
}
var mutex = new(sync.Mutex)

var out = os.Stdout

type buffers [][]byte

func (v *buffers) consume(n int64) {
	for len(*v) > 0 {
		ln0 := int64(len((*v)[0]))
		if ln0 > n {
			(*v)[0] = (*v)[0][n:]
			return
		}
		n -= ln0
		*v = (*v)[1:]
	}
}

func (v *buffers) WriteTo(w io.Writer) (n int64, err error) {
	for _, b := range *v {
		nb, err := w.Write(b)
		n += int64(nb)
		if err != nil {
			v.consume(n)
			return n, err
		}
	}
	v.consume(n)
	return n, nil
}

// write any stored buffers, plus the given line, then empty out
// the buffers.
func (l *clogger) writeBuffers(line []byte) {
	mutex.Lock()
	fmt.Fprintf(out, "\x1b[%dm", colors[l.idx])
	if FlogTime {
		now := time.Now().Format("15:04:05")
		fmt.Fprintf(out, "%s %*s | ", now, maxProcNameLength, l.name)
	} else {
		fmt.Fprintf(out, "%*s | ", maxProcNameLength, l.name)
	}
	fmt.Fprintf(out, "\x1b[m")
	l.buffers = append(l.buffers, line)
	l.buffers.WriteTo(out)
	l.buffers = l.buffers[0:0]
	mutex.Unlock()
}

// bundle writes into lines, waiting briefly for completion of lines
func (l *clogger) writeLines() {
	var tick <-chan time.Time
	for {
		select {
		case w, ok := <-l.writes:
			if !ok {
				if len(l.buffers) > 0 {
					l.writeBuffers([]byte("\n"))
				}
				return
			}
			buf := bytes.NewBuffer(w)
			for {
				line, err := buf.ReadBytes('\n')
				if len(line) > 0 {
					if line[len(line)-1] == '\n' {
						// any text followed by a newline should flush
						// existing buffers. a bare newline should flush
						// existing buffers, but only if there are any.
						if len(line) != 1 || len(l.buffers) > 0 {
							l.writeBuffers(line)
						}
						tick = nil
					} else {
						l.buffers = append(l.buffers, line)
						tick = time.After(l.timeout)
					}
				}
				if err != nil {
					break
				}
			}
			l.done <- struct{}{}
		case <-tick:
			if len(l.buffers) > 0 {
				l.writeBuffers([]byte("\n"))
			}
			tick = nil
		}
	}

}

// write handler of logger.
func (l *clogger) Write(p []byte) (int, error) {
	l.writes <- p
	<-l.done
	return len(p), nil
}

// create logger instance.
func createLogger(name string, colorIndex, colorShift int) *clogger {
	mutex.Lock()
	defer mutex.Unlock()
	if colorShift >= len(colors) {
		colorShift = int(math.Mod(float64(colorShift), float64(len(colors))))
	}
	l := &clogger{idx: colorIndex, name: name, writes: make(chan []byte), done: make(chan struct{}), timeout: 2 * time.Millisecond}
	go l.writeLines()
	return l
}
