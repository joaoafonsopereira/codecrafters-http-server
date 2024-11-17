package myhttp

import (
	"bytes"
	"compress/gzip"
	"io"
	"strconv"
)

type ResponseWriter interface {
	Header() Headers
	Write([]byte) (int, error)
	WriteStatusLine(statusLine []byte)

	serialize() []byte
}

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

func (r *Response) Header() Headers {
	return r.Headers
}

func (r *Response) WriteStatusLine(statusLine []byte) {
	r.statusLine = statusLine
}

func (r *Response) Write(body []byte) (int, error) {
	r.body = body
	r.Headers.Set("Content-Length", strconv.Itoa(len(body)))
	return len(body), nil
}

func (r *Response) WriteTextBody(body []byte) (int, error) {
	r.Headers.Set("Content-Type", "text/plain")
	return r.Write(body)
}

func (r *Response) WriteBinaryBody(body []byte) (int, error) {
	r.Headers.Set("Content-Type", "application/octet-stream")
	return r.Write(body)
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

type EncodedResponse struct {
	Response
	compressionScheme string
}

func NewEncodedResponse(compressionScheme string) *EncodedResponse {
	response := NewResponse()
	response.Headers["Content-Encoding"] = compressionScheme
	return &EncodedResponse{
		Response:          *response,
		compressionScheme: compressionScheme,
	}
}

func (r *EncodedResponse) Write(data []byte) (int, error) {
	var buf bytes.Buffer
	var zw io.WriteCloser

	switch r.compressionScheme {
	case "gzip":
		zw = gzip.NewWriter(&buf)
	default:
		panic("Unknown compression scheme: " + r.compressionScheme)
	}

	n, err := zw.Write(data)
	if err != nil {
		return 0, err
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}

	r.Headers.Set("Content-Length", strconv.Itoa(n))
	return n, nil
}
