package util

import (
	"bufio"
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
)

//GetInputChannel reads identifiers from a file to a channel
func GenInputChannel(ctx context.Context, f *os.File) (int64, <-chan []byte, error) {
	n, err := count(f)
	if err != nil {
		return n, nil, err
	}

	// rewind
	f.Seek(0, io.SeekStart)

	// make the output channel
	identifiers := make(chan []byte)

	// wrap f in a bufio reader
	r := bufio.NewReader(f)
	go func() {
		defer close(identifiers)
		for i := int64(0); i < n; i++ {
			// read next line
			identifier, err := safeReadLine(r)
			if len(identifier) != 0 {
				// push to channel
				identifiers <- identifier
			}
			if err != nil {
				if err != io.EOF {
					zerolog.Ctx(ctx).Error().Err(err).Msg("error reading identifiers: %v")
				}
				return
			}
		}
	}()

	return n, identifiers, nil
}

//count returns number of lines in file
func count(r io.Reader) (int64, error) {
	var n int64
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		n++
	}

	if err := scanner.Err(); err != nil {
		return n, err
	}

	return n, nil
}

// safeReadLine reads each line until a newline character and returns
// read bytes.
func safeReadLine(r *bufio.Reader) (line []byte, err error) {
	// read until newline
	line, err = r.ReadBytes('\n')
	if len(line) > 1 {
		// strip the \n
		line = line[:len(line)-1]
	}
	return
}
