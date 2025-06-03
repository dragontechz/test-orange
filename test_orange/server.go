package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Printf("error in listeningg %v\n", err)
	}

	for {
		conn, err := listener.Accept()
		buffer := make([]byte, 1080)
		if err != nil {
			fmt.Printf("error in accepting func %v \n", err)
		}
		go func() {
			for {
				n, err := conn.Read(buffer)
				if err != nil {
					fmt.Printf("error in reading func %v \n", err)
					conn.Close()
					break
				}

				data := string(buffer[:n])
				fmt.Printf("data recves: %s\n", data)
				conn.Write([]byte("HTTP/1.1 200 Byteok \r\n\r\n"))
			}
		}()
	}
}
