package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

func main() {
	socket, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}

	defer socket.Close()

	for {
		connection, err := socket.Accept()
		if err != nil {
			panic(err)
		}
		go func(connection net.Conn) {
			buffer := make([]byte, 16)
			for {
				length, err := connection.Read(buffer)
				if err != nil {
					return
				}
				if length != 16 {
					fmt.Println("invalid length")
				}

				var ms uint64
				input := bytes.NewReader(buffer[8:])
				err = binary.Read(input, binary.BigEndian, &ms)
				if err != nil {
					panic(err)
				}
				fmt.Printf("got data %v\n", ms)
			}
		}(connection)
	}
}
