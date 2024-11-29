package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var protocol = "HTTP/1.1 200 OK\r\n"

// Define handler functions
func handleRoot(conn net.Conn) error {
	return writeResponse(conn, "HTTP/1.1 200 OK\r\n\r\n")
}

func handleEcho(conn net.Conn, header map[string]string, resBody string) error {
	response := protocol + "Content-Type: " + header["Content-Type: "] + "Content-Length: " + header["Content-Length: "] + resBody
	return writeResponse(conn, response)
}

// userAgentLength
func handleUserAgent(conn net.Conn, header map[string]string, userAgentLength string) error {
	response := protocol + "Content-Type: " + header["Content-Type: "] + "Content-Length: " + userAgentLength + header["User-Agent: "]
	return writeResponse(conn, response)
}

// -------------------------------------------------------- //

// Helper for writing response
func writeResponse(conn net.Conn, response string) error {
	_, err := conn.Write([]byte(response))
	if err != nil {
		return fmt.Errorf("error writing response: %v", err)
	}
	return nil
}

// -------------------------------------------------------- //

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
	// var protocol = "HTTP/1.1 200 OK\r\n"
	var url = strings.TrimSpace(req[1])
	var body = strings.Split(url, "/")
	var resBody = strings.TrimSpace(body[len(body)-1])
	var resBodyLength = len(resBody)

	// First check if there's a User-Agent header before trying to access it
	var userString string
	var userAgentLength string
	if len(req) > 4 {
		userAgentString := req[4:]
		userString = strings.TrimSpace(userAgentString[0])
		userAgentLength = strconv.Itoa(len(userString)) + "\r\n\r\n"
	} else {
		userString = ""
		userAgentLength = "0"
	}

	// Your headers map can stay the same
	var header = map[string]string{
		"Content-Type: ":   "text/plain\r\n",
		"Content-Length: ": strconv.Itoa(resBodyLength) + "\r\n\r\n",
		"User-Agent: ":     userString,
	}

	// routes
	switch {
	case req[1] == "/":
		err = handleRoot(conn)
	case strings.Contains(req[1], "/echo"):
		err = handleEcho(conn, header, resBody)
	case strings.Contains(req[1], "/user-agent"):
		err = handleUserAgent(conn, header, userAgentLength)
	default:
		err = writeResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}

	if err != nil {
		log.Println(err)
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
