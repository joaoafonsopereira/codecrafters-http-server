package myhttp

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
)

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

	response := router.handleRequest(request)

	_, err = conn.Write(response.serialize())
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
