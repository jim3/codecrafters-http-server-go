package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var protocol = "HTTP/1.1 200 OK\r\n"
var http201 = "HTTP/1.1 201 Created\r\n\r\n"

type ParserConfig struct {
	conn     net.Conn
	filePath string
}

func writeResponse(conn net.Conn, response string) error {
	_, err := conn.Write([]byte(response))
	if err != nil {
		return fmt.Errorf("error writing response: %v", err)
	}
	return nil
}

// handler functions
func root(conn net.Conn) error {
	return writeResponse(conn, "HTTP/1.1 200 OK\r\n\r\n")
}

func echo(conn net.Conn, header map[string]string, resBody string) error {
	response := protocol + "Content-Type: " + header["Content-Type: "] + "Content-Length: " + header["Content-Length: "] + resBody
	return writeResponse(conn, response)
}

func userAgent(conn net.Conn, req []string) error {
	var userString string
	var userAgentLength string
	if len(req) > 4 {
		userAgentString := req[4:]
		userString = strings.TrimSpace(userAgentString[0])
		userAgentLength = strconv.Itoa(len(userString))
	} else {
		userString = ""
		userAgentLength = "0"
	}

	// header
	header := map[string]string{
		"Content-Type: ": "text/plain\r\n",
	}

	// send response
	response := protocol + "Content-Type: " + header["Content-Type: "] + "Content-Length: " + userAgentLength + "\r\n\r\n" + userString
	return writeResponse(conn, response)
}

func fileReaderPost(conn net.Conn, filePath string, resBody string, req []string) error {
	str := strings.Join(req, "\n")

	// lines with single \r\n won't match
	re, writeErr := regexp.Compile(`\r\n\r\n[\w\s]+`)
	if writeErr != nil {
		log.Fatal(writeErr)
	}

	match := re.FindAllString(str, -1)
	resStr := strings.Join(match, " ")
	replacedStr := strings.ReplaceAll(resStr, "\n", " ")
	responseString := strings.TrimSpace(replacedStr)

	var fp = filePath + resBody
	if _, err := os.Stat(fp); err == nil {
		return writeResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}

	// Create and write file in one step
	err := os.WriteFile(fp, []byte(responseString), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return writeResponse(conn, http201)
}

func fileReader(conn net.Conn, filePath string, resBody string, req []string) error {
	var fp = filePath + resBody

	if req[0] == "POST" {
		return fileReaderPost(conn, filePath, resBody, req)
	}

	// check if file exists
	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		response := "HTTP/1.1 404 Not Found\r\n\r\n"
		return writeResponse(conn, response)
	}

	f, err := os.ReadFile(fp)
	if err != nil {
		log.Println("Cannot read file:", err)
		return err
	}

	var fileContentLength = strconv.Itoa(len(string(f)))

	header := map[string]string{
		"Content-Type: ": "application/octet-stream\r\n",
	}

	response := protocol + "Content-Type: " + header["Content-Type: "] + "Content-Length: " + fileContentLength + "\r\n\r\n" + string(f)
	return writeResponse(conn, response)
}

// -------------------------------------------------------- //

func httpParser(config ParserConfig) {
	buf := make([]byte, 1024)
	n, err := config.conn.Read(buf) // conn->Read->buf
	if err != nil {
		log.Fatal(err)
	}
	defer config.conn.Close()

	// slice up request
	req := strings.Split(string(buf[:n]), " ") // conn->Read->buf->req

	// request line
	var url = strings.TrimSpace(req[1])
	var body = strings.Split(url, "/")
	var resBody = strings.TrimSpace(body[len(body)-1])
	var resBodyLength = len(resBody)

	// headers
	var header = map[string]string{
		"Content-Type: ":   "text/plain\r\n",
		"Content-Length: ": strconv.Itoa(resBodyLength) + "\r\n\r\n",
	}

	// routes
	switch {
	case req[1] == "/":
		err = root(config.conn)
	case strings.Contains(req[1], "/echo"):
		err = echo(config.conn, header, resBody)
	case strings.Contains(req[1], "/user-agent"):
		err = userAgent(config.conn, req)
	case strings.Contains(req[1], "/files"):
		err = fileReader(config.conn, config.filePath, resBody, req)
	default:
		err = writeResponse(config.conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}

	if err != nil {
		log.Println(err)
	}
}

// -------------------------------------------------------- //

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	// optional file path from args for stage #AP6
	var filePath string
	if len(os.Args) > 2 {
		filePath = os.Args[2]
		fmt.Println("filePath: ", filePath)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		config := ParserConfig{
			conn:     conn,
			filePath: filePath,
		}
		go httpParser(config)
	}
}
