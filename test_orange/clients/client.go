package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":8080")

	if err != nil {
		fmt.Printf(" error in  in dialing connection %v\n", err)
	}

	buffer := make([]byte, 1080)

	for {

		conn.Write([]byte("HTTP/1.1 200 Byteok \r\n\r\n"))

		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("error in recving %v\n", err)
			conn.Close()
			break
		}
		data := string(buffer[:n])
		fmt.Printf("recv: %s\n", data)
	}
}
