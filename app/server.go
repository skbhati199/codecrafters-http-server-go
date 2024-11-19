package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
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

	// Extract the path from the request line
	path := extractPath(requestLine)

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

	// Prepare the HTTP response based on the path
	var response string
	if path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(path, "/echo/") {
		echoStr := strings.TrimPrefix(path, "/echo/")
		contentLength := len(echoStr)
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, echoStr)
	} else if path == "/user-agent" {
		userAgent := headers["user-agent"]
		contentLength := len(userAgent)
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, userAgent)
	} else if strings.HasPrefix(path, "/files/") {
		filename := strings.TrimPrefix(path, "/files/")
		filePath := filepath.Join(directory, filename)
		
		file, err := os.Open(filePath)
		if err != nil {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			defer file.Close()
			
			fileInfo, err := file.Stat()
			if err != nil {
				response = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
			} else {
				contentLength := fileInfo.Size()
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", contentLength)
				_, err = conn.Write([]byte(response)) // Write headers first
				if err != nil {
					fmt.Println("Error writing response:", err.Error())
					return
				}
				// Send file content after successful header write
				_, err = io.Copy(conn, file)
				if err != nil {
					fmt.Println("Error writing file content:", err.Error())
				}
				return // Exit after sending complete response
			}
		}
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	// Write the response
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err.Error())
		return
	}
}

func extractPath(requestLine string) string {
	parts := strings.Split(requestLine, " ")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}