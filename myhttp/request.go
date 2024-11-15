package myhttp

import (
	"bytes"
	"fmt"
)

type Request struct {
	requestLine []byte
	Headers     []byte
	Body        []byte

	Method        string
	Path          string
	methodAndPath string

	PathVariables map[string]string
}

func parseHttpRequest(data []byte) *Request {
	endOfReqLine := bytes.Index(data, []byte("\r\n")) + 1     // idx of \n
	endOfHeaders := bytes.LastIndex(data, []byte("\r\n")) + 1 // idx of \n

	res := &Request{}

	res.requestLine = data[:endOfReqLine+1] // todo include \r\n on line or not?
	res.Headers = data[endOfReqLine+1 : endOfHeaders+1]
	res.Body = data[endOfHeaders+1:]

	method, path, _ := parseRequestLine(res.requestLine)
	res.Method = string(method)
	res.Path = string(path)
	res.methodAndPath = fmt.Sprintf("%s %s", res.Method, res.Path)

	return res
}

func parseRequestLine(line []byte) (method, path, version []byte) {
	components := bytes.Split(line, []byte(" "))
	method = components[0]
	path = components[1]
	version = components[2]
	return
}
