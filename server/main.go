package main

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

		buffer, id := make([]byte, 1), 0
		length, err := connection.Read(buffer)
		if err != nil {
			return
		}
		if length != 1 {
			fmt.Println("invalid length for id")
		}
		dir := fmt.Sprintf("%d", buffer[0])
		entries, err := ioutil.ReadDir(dir)
		if err != nil {
			err = os.Mkdir(dir, 0777)
			if err != nil {
				panic(err)
			}
		} else {
			name := entries[len(entries)-1].Name()
			parts := strings.Split(name, ".")
			id, err = strconv.Atoi(parts[0])
			if err != nil {
				panic(err)
			}
			id++
		}

		out, err := os.Create(fmt.Sprintf("%s/%06d.geiger", dir, id))
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
					fmt.Println("invalid length for time value")
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
