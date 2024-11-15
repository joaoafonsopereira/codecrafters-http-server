package myhttp

import (
	"bytes"
	"strconv"
)

type Response struct {
	statusLine []byte
	Headers    Headers
	body       []byte
}

func NewResponse() *Response {
	return &Response{
		Headers: make(Headers),
	}
}

func (r *Response) WithStatusLine(statusLine []byte) *Response {
	r.statusLine = statusLine
	return r
}

func (r *Response) WithHeader(header, value string) *Response {
	r.Headers[header] = value
	return r
}

func (r *Response) WithBody(body []byte) *Response {
	r.body = body
	return r
}

func (r *Response) WithTextBody(body []byte) *Response {
	return r.
		WithHeader("Content-Type", "text/plain").
		WithHeader("Content-Length", strconv.Itoa(len(body))).
		WithBody(body)
}

func (r *Response) WithBinaryBody(body []byte) *Response {
	return r.
		WithHeader("Content-Type", "application/octet-stream").
		WithHeader("Content-Length", strconv.Itoa(len(body))).
		WithBody(body)
}

func (r *Response) serialize() []byte {
	res := new(bytes.Buffer)
	res.Write(r.statusLine)
	res.Write([]byte("\r\n"))
	res.Write(serializeHeaders(r.Headers))
	res.Write([]byte("\r\n"))
	res.Write(r.body)
	return res.Bytes()
}
