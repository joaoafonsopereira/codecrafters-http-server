package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

var (
	status200 = []byte("HTTP/1.1 200 OK")
	status404 = []byte("HTTP/1.1 404 Not Found")
)

func main() {
	directory := flag.String("--directory", "", "Pass the directory to mount")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn, *directory)
	}

}

func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()

	data, err := readAllData(conn)
	if err != nil {
		fmt.Println("Error reading connection data: ", err.Error())
		os.Exit(1)
	}

	requestLine, headers, _ := parseHttpRequest(data)
	_, path, _ := parseRequestLine(requestLine)

	response := &Response{}

	// routing / handlers
	if bytes.Equal(path, []byte("/")) {
		response = response.withStatusLine(status200)
	} else if bytes.HasPrefix(path, []byte("/echo/")) {
		str, _ := bytes.CutPrefix(path, []byte("/echo/"))

		response = response.
			withStatusLine(status200).
			withTextBody(str)

	} else if bytes.Equal(path, []byte("/user-agent")) {
		userAgent, _ := headerValue(headers, []byte("User-Agent")) // todo assumes header is always present
		response = response.
			withStatusLine(status200).
			withTextBody(userAgent)
	} else if bytes.HasPrefix(path, []byte("/files/")) {
		filename, _ := bytes.CutPrefix(path, []byte("/files/"))
		file := filepath.Join(directory, string(filename))

		content, err := os.ReadFile(file)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Println("Error opening file: ", err.Error())
				os.Exit(1)
			}
			response = response.withStatusLine(status404)
		} else {
			response = response.
				withStatusLine(status200).
				withBinaryBody(content)
		}

	} else {
		response = response.withStatusLine(status404)
	}

	_, err = conn.Write(response.serialize())
	if err != nil {
		fmt.Println("Error writing answer: ", err.Error())
		os.Exit(1)
	}
}

func headerValue(headers []byte, header []byte) (value []byte, found bool) {
	headerStart := bytes.Index(headers, header)
	if headerStart == -1 {
		return nil, false
	}
	valueStart := headerStart + len(header) + 2 // 2 accounts for ": "
	valueEnd := valueStart + bytes.Index(headers[valueStart:], []byte("\r\n"))
	return headers[valueStart:valueEnd], true
}

type Response struct {
	statusLine []byte
	headers    []byte
	body       []byte
}

func (r *Response) withStatusLine(statusLine []byte) *Response {
	r.statusLine = statusLine
	return r
}

func (r *Response) withHeader(header []byte) *Response {
	r.headers = append(r.headers, header...)
	return r
}

func (r *Response) withBody(body []byte) *Response {
	r.body = body
	return r
}

func (r *Response) withTextBody(body []byte) *Response {
	return r.
		withHeader([]byte("Content-Type: text/plain\r\n")).
		withHeader([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body)))).
		withBody(body)
}

func (r *Response) withBinaryBody(body []byte) *Response {
	return r.
		withHeader([]byte("Content-Type: application/octet-stream\r\n")).
		withHeader([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body)))).
		withBody(body)
}

func (r *Response) serialize() []byte {
	res := new(bytes.Buffer)
	res.Write(r.statusLine)
	res.Write([]byte("\r\n"))
	res.Write(r.headers)
	res.Write([]byte("\r\n"))
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

func parseHttpRequest(data []byte) (requestLine, headers, body []byte) {
	endOfReqLine := bytes.Index(data, []byte("\r\n")) + 1     // idx of \n
	endOfHeaders := bytes.LastIndex(data, []byte("\r\n")) + 1 // idx of \n

	requestLine = data[:endOfReqLine+1] // todo include \r\n on line or not?
	headers = data[endOfReqLine+1 : endOfHeaders+1]
	body = data[endOfHeaders+1:]
	return
}

func parseRequestLine(line []byte) (method, path, version []byte) {
	components := bytes.Split(line, []byte(" "))
	method = components[0]
	path = components[1]
	version = components[2]
	return
}
