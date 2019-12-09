package main

import (
	"fmt"
	"math/big"
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

				var ms big.Int
				ms.SetBytes(buffer)
				fmt.Printf("got data %v\n", ms.String())
			}
		}(connection)
	}
}
