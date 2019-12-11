package main

import (
	"compress/gzip"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	socket, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}

	defer socket.Close()

	type File struct {
		file   *os.File
		writer *gzip.Writer
	}
	lock, files, index := sync.Mutex{}, make(map[uint64]File), uint64(0)
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		lock.Lock()
		defer lock.Unlock()
		for _, file := range files {
			file.writer.Close()
			file.file.Close()
		}
		os.Exit(0)
	}()
	for {
		connection, err := socket.Accept()
		if err != nil {
			panic(err)
		}
		out, err := os.Create(fmt.Sprintf("data_%d.geiger", index))
		if err != nil {
			panic(err)
		}
		file := File{
			file:   out,
			writer: gzip.NewWriter(out),
		}
		lock.Lock()
		files[index] = file
		index++
		lock.Unlock()
		go func(index uint64, connection net.Conn, out File) {
			defer func() {
				lock.Lock()
				delete(files, index)
				out.writer.Close()
				out.file.Close()
				lock.Unlock()
			}()
			buffer, last := make([]byte, 16), big.NewInt(0)
			for {
				length, err := connection.Read(buffer)
				if err != nil {
					return
				}
				if length != 16 {
					fmt.Println("invalid length")
				}

				ms := big.NewInt(0)
				ms.SetBytes(buffer)
				fmt.Printf("got data %v\n", ms.String())
				diff := big.NewInt(0)
				diff.Sub(ms, last)
				encoded := diff.Bytes()
				if size := 16 - len(encoded); size > 0 {
					padding := make([]byte, size)
					_, err := out.writer.Write(padding)
					if err != nil {
						return
					}
				}
				_, err = out.writer.Write(encoded)
				if err != nil {
					return
				}
				last = ms
			}
		}(index, connection, file)
	}
}
