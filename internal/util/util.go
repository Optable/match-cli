/*
Created on Feb 24, 2021

@author: Akshaya Mani, Optable Inc.
*/
package util

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog"
)

func NewTLS(caFile string, certFile string, keyFile string, serverName string) (*tls.Config, error) {
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("Failed to parse ca file")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:      cp,
		ClientCAs:    cp,
		Certificates: []tls.Certificate{cert},
		ServerName:   serverName,
		// Only used on the server side but noop on the client
		ClientAuth: tls.RequireAndVerifyClientCert,
	}, nil
}

//Read input from a file into a slice
func ReadInput(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	var input []string

	for scanner.Scan() {
		input = append(input, scanner.Text())
	}

	return input, scanner.Err()
}

//Write output from a slice into a file
func WriteOutput(output []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()
	w := bufio.NewWriter(file)

	for _, line := range output {
		fmt.Fprintln(w, line)
	}

	return w.Flush()
}

//Read identifiers from a file to a channel
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
	src := bufio.NewReader(f)
	go func() {
		defer close(identifiers)
		for i := int64(0); i < n; i++ {
			identifier, err := SafeReadLine(src)
			if len(identifier) != 0 {
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

func SafeReadLine(r *bufio.Reader) (line []byte, err error) {
	// read until newline
	line, err = r.ReadBytes('\n')
	if len(line) > 1 {
		// strip the \n
		line = line[:len(line)-1]
	}
	return
}
