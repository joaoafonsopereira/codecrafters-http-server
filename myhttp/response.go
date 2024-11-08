package myhttp

import (
	"bytes"
	"fmt"
)

type Response struct {
	statusLine []byte
	headers    []byte
	body       []byte
}

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) WithStatusLine(statusLine []byte) *Response {
	r.statusLine = statusLine
	return r
}

func (r *Response) WithHeader(header []byte) *Response {
	r.headers = append(r.headers, header...)
	return r
}

func (r *Response) WithBody(body []byte) *Response {
	r.body = body
	return r
}

func (r *Response) WithTextBody(body []byte) *Response {
	return r.
		WithHeader([]byte("Content-Type: text/plain\r\n")).
		WithHeader([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body)))).
		WithBody(body)
}

func (r *Response) WithBinaryBody(body []byte) *Response {
	return r.
		WithHeader([]byte("Content-Type: application/octet-stream\r\n")).
		WithHeader([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body)))).
		WithBody(body)
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
