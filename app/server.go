package main

import (
	"fmt"
	"net"
	"os"
)

// strings.Split(" ")
func packetParser(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Processing connection:", conn.RemoteAddr())

	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	fmt.Println("Response sent to client")
}

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		// Wait for connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently
		go packetParser(conn)
	}
}

// -------------------------------------------------------------
// HTTP/1.1 200 OK\r\n\r\n
// line, err := reader.ReadString()
// fmt.Println("data := bufio.NewReader(conn): ", data)
