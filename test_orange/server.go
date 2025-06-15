package main

import (
	"fmt"
	"log"
	"net"
	"stream/utils"
)

const (
	start    = "Bytes1"
	end      = "Bytes0"
	BUFFSIZE = 1080 * 50
	mask     = "HTTP/1.1 200 Bytes1%sBytes0\r\n\r\n"
)

func main() {
	listener, err := net.Listen("tcp", ":9090")

	if err != nil {
		fmt.Printf("error in listeningg %v\n", err)
	}
	proxy := utils.SOCKS5{Listn_addr: ":8080"}
	go proxy.RUN_v5()

	for {
		conn, err := listener.Accept()
		//buffer := make([]byte, 1080)
		if err != nil {
			fmt.Printf("error in accepting func %v \n", err)
		}
		conn2, err := net.Dial("tcp", "127.0.0.1:8080")

		if err != nil {
			fmt.Printf(" error in  in dialing connection %v\n", err)
		}
		fmt.Printf("handling client\n")
		go handle_stream(conn, conn2)
	}
}

func handle_stream(src, dst net.Conn) {
	req_channel := make(chan string)
	resp_channel := make(chan string)
	go read_resp(dst, resp_channel)
	go read_req(src, req_channel)
	for {
		go forward(src, resp_channel)
		forward(dst, req_channel)
		break
	}

}

func read_req(conn net.Conn, channel chan string) {
	buffer := make([]byte, BUFFSIZE)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("error in recving %v\n", err)
			conn.Close()
			close(channel)
			break
		}
		conn_read := string(buffer[:n])
		fmt.Println("from read req: ", conn_read)

		data, err := utils.GET(conn_read, start, end)
		if err != nil {
			log.Printf("Erreur: %v", err)
		}

		channel <- data

	}
}
func read_resp(conn net.Conn, channel chan string) {
	buffer := make([]byte, BUFFSIZE)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("error in recving %v\n", err)
			conn.Close()
			close(channel)
			break
		}
		conn_read := string(buffer[:n])

		data := fmt.Sprintf(mask, conn_read)
		fmt.Printf("fowarding data to the client : %s\n", data)
		channel <- data

	}
}

func forward(conn net.Conn, channel chan string) {
	for {
		data, ok := <-channel

		if !ok {
			conn.Close()
			break
		}

		_, err := conn.Write([]byte(data))
		if err != nil {
			fmt.Printf("error in reading from channel %v\n", err)
			conn.Close()
			break
		}
	}
}
