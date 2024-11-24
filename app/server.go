package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func packetParser(conn net.Conn) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf) // conn<-Read<-buf
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := strings.Split(string(buf[:n]), " ")
	if req[1] != "/" {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
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
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go packetParser(conn)
	}
}
