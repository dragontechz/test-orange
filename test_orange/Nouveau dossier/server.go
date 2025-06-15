package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

const (
	LISTEN_PORT = ":9090"
	PROXY_ADDR  = "127.0.0.1:8080"
	KEY         = "4d7953757065725365637265744b6579466f7254657374696e67313233" // Même clé !
)

func main() {
	key, _ := hex.DecodeString(KEY)
	listener, err := net.Listen("tcp", LISTEN_PORT)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Printf("Client proxy démarré sur %s\n", LISTEN_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Erreur accept:", err)
			continue
		}
		go handleClient(conn, key)
	}
}

func handleClient(clientConn net.Conn, key []byte) {
	defer clientConn.Close()

	proxyConn, err := net.Dial("tcp", PROXY_ADDR)
	if err != nil {
		log.Println("Erreur connexion proxy:", err)
		return
	}
	defer proxyConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Proxy
	go func() {
		defer wg.Done()
		transferHTTPEncapsulated(clientConn, proxyConn, key)
	}()

	// Proxy -> Client
	go func() {
		defer wg.Done()
		transferEncrypted(proxyConn, clientConn, key, "Bytes1", "Bytes0")
	}()

	wg.Wait()
}
