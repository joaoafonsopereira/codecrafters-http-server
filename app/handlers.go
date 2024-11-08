package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

func downloadHandler(path []byte, directory string) *Response {
	response := &Response{}

	filename, _ := bytes.CutPrefix(path, []byte("/files/"))
	file := filepath.Join(directory, string(filename))

	content, err := os.ReadFile(file)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error opening file: ", err.Error())
			os.Exit(1)
		}
		return response.withStatusLine(status404)
	}

	return response.
		withStatusLine(status200).
		withBinaryBody(content)
}

func uploadHandler(path, body []byte, directory string) *Response {
	response := &Response{}

	filename, _ := bytes.CutPrefix(path, []byte("/files/"))
	file := filepath.Join(directory, string(filename))

	err := os.WriteFile(file, body, 0644)
	if err != nil {
		fmt.Println("Error writing file: ", err.Error())
		os.Exit(1)
	}

	return response.
		withStatusLine(status201)
}
