package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/myhttp"
	"net"
	"os"
	"path/filepath"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	directory := flag.String("directory", "", "Pass the directory to mount")
	flag.Parse()

	router := myhttp.NewRouter()

	router.RegisterHandler("/", func(_ *myhttp.Request) *myhttp.Response {
		return myhttp.NewResponse().WithStatusLine(myhttp.Status200)
	})

	router.RegisterHandler("/echo", func(req *myhttp.Request) *myhttp.Response {
		str, _ := bytes.CutPrefix([]byte(req.Path), []byte("/echo/"))

		return myhttp.NewResponse().
			WithStatusLine(myhttp.Status200).
			WithTextBody(str)
	})

	router.RegisterHandler("/user-agent", func(req *myhttp.Request) *myhttp.Response {
		userAgent, _ := headerValue(req.Headers, []byte("User-Agent")) // todo assumes header is always present
		return myhttp.NewResponse().
			WithStatusLine(myhttp.Status200).
			WithTextBody(userAgent)
	})

	router.RegisterHandler("GET /files/", func(req *myhttp.Request) *myhttp.Response {
		return downloadHandler([]byte(req.Path), *directory)
	})

	router.RegisterHandler("POST /files/", func(req *myhttp.Request) *myhttp.Response {
		return uploadHandler([]byte(req.Path), req.Body, *directory)
	})

	err := myhttp.ListenAndServe("tcp", "0.0.0.0:4221", router)
	if err != nil {
		fmt.Println("Failed to start server: " + err.Error())
		os.Exit(1)
	}

}

// todo this should go to myhttp package
func headerValue(headers []byte, header []byte) (value []byte, found bool) {
	headerStart := bytes.Index(headers, header)
	if headerStart == -1 {
		return nil, false
	}
	valueStart := headerStart + len(header) + 2 // 2 accounts for ": "
	valueEnd := valueStart + bytes.Index(headers[valueStart:], []byte("\r\n"))
	return headers[valueStart:valueEnd], true
}

func downloadHandler(path []byte, directory string) *myhttp.Response {
	response := myhttp.NewResponse()

	filename, _ := bytes.CutPrefix(path, []byte("/files/"))
	file := filepath.Join(directory, string(filename))

	content, err := os.ReadFile(file)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error opening file: ", err.Error())
			os.Exit(1)
		}
		return response.WithStatusLine(myhttp.Status404)
	}

	return response.
		WithStatusLine(myhttp.Status200).
		WithBinaryBody(content)
}

func uploadHandler(path, body []byte, directory string) *myhttp.Response {
	response := myhttp.NewResponse()

	filename, _ := bytes.CutPrefix(path, []byte("/files/"))
	file := filepath.Join(directory, string(filename))

	err := os.WriteFile(file, body, 0644)
	if err != nil {
		fmt.Println("Error writing file: ", err.Error())
		os.Exit(1)
	}

	return response.
		WithStatusLine(myhttp.Status201)
}
