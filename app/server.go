package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func httpParser(conn net.Conn) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf) // conn -> Read -> buf
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// slice up request
	req := strings.Split(string(buf[:n]), " ")

	// request line
	var protocol = "HTTP/1.1 200 OK\r\n"
	var url = strings.TrimSpace(req[1])
	var body = strings.Split(url, "/")
	var resBody = strings.TrimSpace(body[len(body)-1])
	var contentLength = len(resBody)

	// headers
	var header = map[string]string{
		"Content-Type: ":   "text/plain\r\n",
		"Content-Length: ": strconv.Itoa(contentLength),
	}

	// routes
	switch {
	case req[1] == "/":
		_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if err != nil {
			log.Println("Error writing response:", err)
		}
	case strings.Contains(req[1], "/echo"):
		_, err := conn.Write([]byte(protocol + "Content-Type: " + header["Content-Type: "] + "Content-Length: " + header["Content-Length: "] + "\r\n\r\n" + resBody))
		if err != nil {
			log.Println("Error writing response:", err)
		}
	default:
		_, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		if err != nil {
			log.Println("Error writing response:", err)
		}
	}
}

// --------------------------------------------------------------

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go httpParser(conn)
	}
}
