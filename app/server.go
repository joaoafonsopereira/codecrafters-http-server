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

var (
	status200 = []byte("HTTP/1.1 200 OK\r\n\r\n")
	status404 = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
)

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
	var response Response

	if bytes.Equal(path, []byte("/")) {
		response.statusLine = status200
	} else if bytes.HasPrefix(path, []byte("/echo")) {
		//str, hasPrefix := bytes.CutPrefix(path, []byte("/echo/")) //todo handle wrong prefix case
		str, _ := bytes.CutPrefix(path, []byte("/echo/"))
		headers := new(bytes.Buffer)
		headers.WriteString("Content-Type: text/plain\r\n")
		headers.WriteString(
			fmt.Sprintf("Content-Length: %d\r\n", len(str)),
		)
		response.statusLine = status200
		response.headers = headers.Bytes()
		response.body = str
	} else {
		response.statusLine = status404
	}

	_, err = conn.Write(response.serialize())
	if err != nil {
		fmt.Println("Error writing answer: ", err.Error())
		os.Exit(1)
	}
}

type Response struct {
	statusLine []byte
	headers    []byte
	body       []byte
}

func (r *Response) serialize() []byte {
	res := new(bytes.Buffer)
	res.Write(r.statusLine)
	res.Write(r.headers)
	res.Write(r.body)
	return res.Bytes()
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
