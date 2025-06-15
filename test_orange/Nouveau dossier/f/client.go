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

	// 1. Connexion au serveur proxy
	proxyConn, err := net.Dial("tcp", PROXY_ADDR)
	if err != nil {
		log.Println("Erreur connexion proxy:", err)
		return
	}
	defer proxyConn.Close()

	// 2. Communication bidirectionnelle
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Proxy
	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 4096)
			n, err := clientConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Println("Erreur lecture client:", err)
				}
				break
			}

			// Chiffrement
			encrypted, err := encrypt(buf[:n], key)
			if err != nil {
				log.Println("Erreur chiffrement:", err)
				break
			}

			// Encapsulation HTTP
			httpReq := fmt.Sprintf("GET / HTTP/1.1\r\n\r\nBytes1%sBytes0", hex.EncodeToString(encrypted))
			if _, err := proxyConn.Write([]byte(httpReq)); err != nil {
				log.Println("Erreur envoi proxy:", err)
				break
			}
		}
	}()

	// Proxy -> Client
	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 4096)
			n, err := proxyConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Println("Erreur lecture proxy:", err)
				}
				break
			}

			// Extraction des données
			dataStr := string(buf[:n])
			encrypted := extractData(dataStr)
			if encrypted == nil {
				log.Println("Format invalide")
				break
			}

			// Déchiffrement
			data, err := decrypt(encrypted, key)
			if err != nil {
				log.Println("Erreur déchiffrement:", err)
				break
			}

			// Renvoi au client
			if _, err := clientConn.Write(data); err != nil {
				log.Println("Erreur envoi client:", err)
				break
			}
		}
	}()

	wg.Wait()
}

// [...] (Fonctions extractData, encrypt, decrypt identiques)