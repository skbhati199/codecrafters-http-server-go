package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	directory := flag.String("directory", "", "the directory to serve files from")
	flag.Parse()

	if *directory == "" {
		fmt.Println("Please provide a directory using the --directory flag")
		os.Exit(1)
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go handleConnection(conn, *directory)
	}
}

func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read the request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request:", err.Error())
		return
	}

	method, path := extractMethodAndPath(requestLine)

	// Read headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading headers:", err.Error())
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[strings.ToLower(parts[0])] = parts[1]
		}
	}

	// Handle the request
	switch method {
	case "GET":
		handleGet(conn, directory, path)
	case "POST":
		contentLength, _ := strconv.Atoi(headers["content-length"])
		handlePost(conn, directory, path, reader, contentLength)
	default:
		conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
	}
}

func handleGet(conn net.Conn, directory, path string) {
	filename := strings.TrimPrefix(path, "/files/")
	filePath := filepath.Join(directory, filename)

	file, err := os.Open(filePath)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}

	contentLength := fileInfo.Size()
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", contentLength)
	conn.Write([]byte(response))

	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error writing file content:", err.Error())
	}
}

func handlePost(conn net.Conn, directory, path string, reader *bufio.Reader, contentLength int) {
	filename := strings.TrimPrefix(path, "/files/")
	if filename == "" {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	filePath := filepath.Join(directory, filename)
	file, err := os.Create(filePath)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}
	defer file.Close()

	// Read the request body and write it to the file
	body := make([]byte, contentLength)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}

	_, err = file.Write(body)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}

	conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
}

func extractMethodAndPath(requestLine string) (string, string) {
	parts := strings.Split(requestLine, " ")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
