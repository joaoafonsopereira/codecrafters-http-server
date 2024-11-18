package myhttp

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

var (
	Status200 = []byte("HTTP/1.1 200 OK")
	Status201 = []byte("HTTP/1.1 201 Created")
	Status404 = []byte("HTTP/1.1 404 Not Found")
)

var supportedCompressionSchemes = map[string]bool{
	"gzip": true,
}

func ListenAndServe(network string, addr string, router *Router) error {
	l, err := net.Listen(network, addr)
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

		go handleConnection(conn, router)
	}
}

func handleConnection(conn net.Conn, router *Router) {
	defer conn.Close()

	data, err := readAllData(conn)
	if err != nil {
		fmt.Println("Error reading connection data: ", err.Error())
		os.Exit(1)
	}

	request := parseHttpRequest(data)
	responseWriter := selectResponseWriter(request.Headers)

	handler, context := router.match(request)
	request.PathVariables = context // todo would it make sense to just to this inside router.match?

	if handler == nil {
		responseWriter.WriteStatusLine(Status404)
	} else {
		handler(responseWriter, request)
	}

	_, err = conn.Write(responseWriter.serialize())
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

func selectResponseWriter(headers Headers) ResponseWriter {
	encodingOptions, wantsEncoding := headers["Accept-Encoding"]
	if wantsEncoding {
		scheme := selectEncoding(encodingOptions)
		if scheme != "" {
			return NewEncodedResponse(scheme)
		}
	}
	return NewResponse()
}

func selectEncoding(options string) string {
	opts := strings.Split(options, ",")
	for _, opt := range opts {
		opt = strings.TrimSpace(opt)
		if supportedCompressionSchemes[opt] {
			return opt
		}
	}
	return ""
}
