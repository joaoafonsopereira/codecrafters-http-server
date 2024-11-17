package main

import (
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

	router.RegisterHandler("/", func(rw myhttp.ResponseWriter, _ *myhttp.Request) {
		rw.WriteStatusLine(myhttp.Status200)
	})

	router.RegisterHandler("/echo/{str}", func(rw myhttp.ResponseWriter, req *myhttp.Request) {
		str := req.PathVariables["str"]

		rw.WriteStatusLine(myhttp.Status200)
		rw.WriteTextBody([]byte(str)) // todo maybe api could use strings instead of []byte ?
	})

	router.RegisterHandler("/user-agent", func(rw myhttp.ResponseWriter, req *myhttp.Request) {
		userAgent, _ := req.Headers["User-Agent"] // todo assumes header is always present

		rw.WriteStatusLine(myhttp.Status200)
		rw.WriteTextBody([]byte(userAgent))
	})

	router.RegisterHandler("GET /files/{file}", func(rw myhttp.ResponseWriter, req *myhttp.Request) {
		downloadHandler(rw, req, *directory)
	})

	router.RegisterHandler("POST /files/{file}", func(rw myhttp.ResponseWriter, req *myhttp.Request) {
		uploadHandler(rw, req, *directory)
	})

	err := myhttp.ListenAndServe("tcp", "0.0.0.0:4221", router)
	if err != nil {
		fmt.Println("Failed to start server: " + err.Error())
		os.Exit(1)
	}

}

func downloadHandler(rw myhttp.ResponseWriter, req *myhttp.Request, directory string) {
	filename := req.PathVariables["file"]
	file := filepath.Join(directory, filename)

	content, err := os.ReadFile(file)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error opening file: ", err.Error())
			os.Exit(1)
		}
		rw.WriteStatusLine(myhttp.Status404)
		return
	}

	rw.WriteStatusLine(myhttp.Status200)
	rw.WriteBinaryBody(content)
}

func uploadHandler(rw myhttp.ResponseWriter, req *myhttp.Request, directory string) {
	filename := req.PathVariables["file"]
	file := filepath.Join(directory, filename)

	err := os.WriteFile(file, req.Body, 0644)
	if err != nil {
		fmt.Println("Error writing file: ", err.Error())
		os.Exit(1)
	}

	rw.WriteStatusLine(myhttp.Status201)
}
