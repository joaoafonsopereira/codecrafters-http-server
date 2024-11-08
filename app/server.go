package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	//Uncomment this block to pass the first stage
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	data, err := readAllData(conn)
	if err != nil {
		fmt.Println("Error reading connection data: ", err.Error())
		os.Exit(1)
	}

	requestLine, _, _ := parseHttpRequest(data)
	_, path, _ := parseRequestLine(requestLine)

	// routing / handlers
	var response []byte
	if bytes.Equal(path, []byte("/")) {
		response = []byte("HTTP/1.1 200 OK\r\n\r\n")
	} else {
		response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
	}

	_, err = conn.Write(response)
	if err != nil {
		fmt.Println("Error writing answer: ", err.Error())
		os.Exit(1)
	}

}

func readAllData(conn net.Conn) ([]byte, error) {
	InBufferLen := 1024
	inBuffer := make([]byte, InBufferLen)
	outBuffer := new(bytes.Buffer)

	for {
		n, err := conn.Read(inBuffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err) // todo error msg
				return nil, err
			}
			break
		}
		outBuffer.Write(inBuffer[:n])
		if n < InBufferLen {
			break
		}
	}
	return outBuffer.Bytes(), nil
}

func parseHttpRequest(data []byte) ([]byte, []byte, []byte) {
	endOfReqLine := bytes.Index(data, []byte("\r\n")) + 1     // idx of \n
	endOfHeaders := bytes.LastIndex(data, []byte("\r\n")) + 1 // idx of \n

	requestLine := data[:endOfReqLine+1] // todo include \r\n on line or not?
	headers := data[endOfReqLine+1 : endOfHeaders+1]
	body := data[endOfHeaders+1:]
	return requestLine, headers, body
}

func parseRequestLine(line []byte) (method, path, version []byte) {
	components := bytes.Split(line, []byte(" "))
	method = components[0]
	path = components[1]
	version = components[2]
	return
}
