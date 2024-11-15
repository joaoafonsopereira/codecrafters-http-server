package myhttp

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

func parseHeaders(data []byte) Headers {
	res := make(Headers)
	lines := bytes.Split(data, []byte("\r\n"))
	for _, line := range lines {
		colonIdx := bytes.IndexByte(line, ':')
		headerName := string(line[:colonIdx])
		headerValue := string(line[colonIdx+1:])
		res[headerName] = headerValue
	}
	return res
}

func serializeHeaders(headers Headers) []byte {
	res := new(bytes.Buffer)
	for key, value := range headers {
		line := fmt.Sprintf("%s: %s\r\n", key, value)
		res.WriteString(line)
	}
	return res.Bytes()
}
