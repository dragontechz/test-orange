package utils

import (
	"fmt"
	"io"
	"log"
	"strconv"
	//	"io"
	"net"
	"strings"
)

type SOCKS5 struct {
	Listn_addr string
}
type Proxy struct {
	Listening_Port string
}

func LogRequest(method, url, addr string) {
	log.Printf("Received request: %s %s from %s", method, url, addr)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Lire la requête du client
	requestBuffer := make([]byte, 4096)
	n, err := conn.Read(requestBuffer)
	if err != nil {
		log.Printf("Error reading from connection: %v", err)
		return
	}

	// Convertir la requête en chaîne de caractères
	request := string(requestBuffer[:n])

	// Extraire la ligne de la requête (première ligne)
	var method, url string
	_, err = fmt.Sscanf(request, "%s %s", &method, &url)
	if err != nil {
		log.Printf("Error parsing request: %v", err)
		return
	}

	//logRequest(method, url, conn.RemoteAddr().String())

	if method == "CONNECT" {
		var targetHost string
		_, err = fmt.Sscanf(url, "%s", &targetHost)
		if err != nil {
			log.Printf("Error parsing target host: %v", err)
			return
		}

		// Établir une connexion avec le serveur cible
		targetConn, err := net.Dial("tcp", targetHost) // Utilise l'hôte et le port fournis
		if err != nil {
			log.Printf("Could not connect to target: %v", err)
			return
		}
		defer targetConn.Close()

		// Répondre au client que la connexion est établie
		conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

		// Transférer les données entre le client et le serveur cible
		go io.Copy(targetConn, conn) // Transférer du client vers le serveur
		io.Copy(conn, targetConn)    // Transférer du serveur vers le client
	}

	//method GET etc
	// Extraire l'hôte de l'URL
	// On utilise seulement le nom d'hôte sans le chemin ni les paramètres
	host := url[len("http://"):]

	// Trouver le premier '/' pour couper le chemin
	if pathIdx := strings.Index(host, "/"); pathIdx != -1 {
		host = host[:pathIdx] // Prendre uniquement l'hôte
	}

	// Établir une connexion avec le serveur cible
	if strings.Index(host, ":") == -1 {
		host = host + ":80"
	}
	targetConn, err := net.Dial("tcp", host) // Ajouter le port par défaut si nécessaire
	//	targetConn, err := net.Dial("tcp", host+":80")
	if err != nil {
		log.Printf("Could not connect to target: %v", err)
		return
	}
	defer targetConn.Close()

	// Transférer la requête au serveur cible
	_, err = targetConn.Write(requestBuffer[:n])
	if err != nil {
		log.Printf("Error writing to target connection: %v", err)
		return
	}

	go io.Copy(conn, targetConn)
	io.Copy(targetConn, conn)
}
func (p *Proxy) Run() {
	log.Printf("Proxy HTTP transparent en cours d'exécution sur %s\n", p.Listening_Port)

	listener, err := net.Listen("tcp", p.Listening_Port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		fmt.Println("new connection initialised")
		go handleConnection(conn) // Gérer chaque connexion dans une goroutine
	}
}

const (
	socks5Ver         = 0x05
	cmdConnect        = 0x01
	atypIPv4          = 0x01
	atypDomain        = 0x03
	socks5NoAuth      = 0x00
	socks5Success     = 0x00
	socks5GeneralFail = 0x01
)

func handleConnection_sock(clientConn net.Conn) {
	defer clientConn.Close()

	// Read SOCKS5 greeting
	buf := make([]byte, 256)
	n, err := clientConn.Read(buf)
	if err != nil {
		log.Printf("Error reading greeting: %v", err)
		return
	}

	// Check version and authentication methods
	if n < 2 || buf[0] != socks5Ver || buf[1] == 0 {
		return
	}

	// Respond with no authentication required
	_, err = clientConn.Write([]byte{socks5Ver, socks5NoAuth})
	if err != nil {
		log.Printf("Error sending auth response: %v", err)
		return
	}

	// Read connection request
	n, err = clientConn.Read(buf)
	if err != nil {
		log.Printf("Error reading request: %v", err)
		return
	}
	if n < 7 || buf[0] != socks5Ver || buf[1] != cmdConnect {
		return // Only supporting CONNECT command
	}

	// Parse destination address
	var host string
	var port uint16
	var portStr string

	switch buf[3] { // ATYP field
	case atypIPv4:
		if n < 10 {
			return
		}
		host = net.IPv4(buf[4], buf[5], buf[6], buf[7]).String()
		port = uint16(buf[8])<<8 | uint16(buf[9])
	case atypDomain:
		domainLen := int(buf[4])
		if n < 7+domainLen {
			return
		}
		host = string(buf[5 : 5+domainLen])
		port = uint16(buf[5+domainLen])<<8 | uint16(buf[6+domainLen])
		portStr = strconv.Itoa(int(port))
	default:
		return // Unsupported address type
	}

	// Connect to destination
	//dst_host := net.JoinHostPort(host[:len(host)-1], string(port))
	//fmt.Println("dst host : ", dst_host)
	destConn, err := net.Dial("tcp", host+":"+portStr)
	if err != nil {
		clientConn.Write([]byte{socks5Ver, socks5GeneralFail, 0x00, atypIPv4, 0, 0, 0, 0, 0, 0})
		log.Printf("Error connecting to destination %s:%d: %v", host, port, err)
		return
	}
	defer destConn.Close()

	// Send success response
	resp := []byte{socks5Ver, socks5Success, 0x00, atypIPv4, 0, 0, 0, 0, 0, 0}
	_, err = clientConn.Write(resp)
	if err != nil {
		log.Printf("Error sending success response: %v", err)
		return
	}

	// Bidirectional copy
	go func() {
		_, err := io.Copy(destConn, clientConn)
		if err != nil && err != io.EOF {
			log.Printf("Error copying client to dest: %v", err)
		}
	}()

	_, err = io.Copy(clientConn, destConn)
	if err != nil && err != io.EOF {
		log.Printf("Error copying dest to client: %v", err)
	}
}

func (s *SOCKS5) RUN_v5() {
	listener, err := net.Listen("tcp", s.Listn_addr)
	if err != nil {
		log.Fatalf("Error starting SOCKS5 server: %v", err)
	}
	defer listener.Close()
	log.Printf("SOCKS5 proxy listening on %s", s.Listn_addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		fmt.Printf("new connection\n")
		go handleConnection_sock(conn)
	}
}
