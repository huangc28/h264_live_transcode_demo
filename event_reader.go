package main

import (
	"bufio"
	"bytes"
	"io"

	"github.com/golang/glog"
)

// http://thrawn01.org/posts/2016/10/17/delimited-stream-processing-in-golang/
// https://www.geeksforgeeks.org/how-to-find-the-index-value-of-any-element-in-slice-of-bytes-in-golang/
var NALSeperator []byte = []byte{0x00, 0x00, 0x00, 0x01}

type EventReader struct {
	scanner *bufio.Scanner
}

func IndexSliceByte(bytes []byte, needles []byte) int {
	var i int = -1

	if len(bytes) < len(needles) {
		return -1
	}
OuterLoop:
	for idx, b := range bytes {

		// Try to match first byte in the needles.
		if needles[0] != b {
			i = -1

			continue
		} else {
			i = idx

		InnerLoop:
			for j := 1; j < len(needles); j++ {
				// What if bytes length is less then i+j? That means it's impossible to find the exact `
				// match of `needles` in `bytes` array. We simply break the loop and set the result
				// to be not found.
				if i+j >= len(bytes) {
					i = -1

					break InnerLoop
				}

				if bytes[i+j] != needles[j] {
					i = -1

					break InnerLoop
				}
			}

			// If the next three bytes all matches we simply break the outer loop.
			if i >= 0 {
				break OuterLoop
			}
		}
	}

	return i
}

func NewEventReader(source io.Reader) *EventReader {
	scanner := bufio.NewScanner(source)

	split := func(data []byte, atEOF bool) (int, []byte, error) {
		// If we are at the end of file and we have no data in the current
		// data buffer, we request for more data.
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// If we found NALSeperator in the current data buffer we have a full event.
		// we can flush the data out of the buffer.
		if i := IndexSliceByte(data, NALSeperator); i >= 0 {
			return i + 1, data[0:i], nil
		}

		// If we're at EOF, we have a final event
		if atEOF {
			return len(data), data, nil
		}

		// Request for more data.
		return 0, nil, nil
	}

	scanner.Split(split)

	return &EventReader{
		scanner: scanner,
	}
}

func (r *EventReader) ReadEvent() io.ReadCloser {
	// There might be more than 1 frames, thus, we need to
	// iterate through each frame.
	pr, pw := io.Pipe()

	go func() {
		for r.scanner.Scan() {
			// Minimum NAL unit.
			frame := r.scanner.Bytes()

			if err := r.scanner.Err(); err != nil {
				glog.Errorf("DEBUG failed to scan frame %v", err)

				return
			}

			// Initialize byte reader using frame bytes.
			// If is first frame, append 0x0, 0x0, 0x0, 0x1.
			frame = append(NALSeperator, frame...)

			fr := bytes.NewReader(frame)

			io.Copy(pw, fr)
		}

		pw.Close()
	}()

	return pr
}
