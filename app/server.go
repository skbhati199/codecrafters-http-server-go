package main
import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)
const CRLF = "\r\n"
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	dir := flag.String("directory", "", "enter a directory")
	flag.Parse()
	ln, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221", err)
		os.Exit(1)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			os.Exit(1)
		}
		go func() {
			defer conn.Close()
			buf := make([]byte, 1024)
			_, err = conn.Read(buf)
			if err != nil {
				fmt.Println("Error accepting connection: ", err)
			}
			req := string(buf)
			lines := strings.Split(req, CRLF)
			path := strings.Split(lines[0], " ")[1]
			method := strings.Split(lines[0], " ")[0]
			fmt.Println(path)
			var res string
			if path == "/" {
				res = "HTTP/1.1 200 OK\r\n\r\n"
			} else if strings.HasPrefix(path, "/echo/") {
				msg := path[6:]
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(msg), msg)
			} else if path == "/user-agent" {
				msg := strings.Split(lines[2], " ")[1]
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(msg), msg)
			} else if strings.HasPrefix(path, "/files/") && *dir != "" {
				filename := path[7:]
				fmt.Println(*dir + filename)
				if method == "GET" {
					if file, err := os.ReadFile(*dir + filename); err == nil {
						content := string(file)
						res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
					} else {
						res = "HTTP/1.1 404 Not found\r\n\r\n"
					}
				} else if method == "POST" {
					var contentLength int
					for _, line := range lines {
						if strings.HasPrefix(strings.ToLower(line), "content-length:") {
							fmt.Sscanf(line, "Content-Length: %d", &contentLength)
							break
						}
					}

					if contentLength <= 0 {
						res = "HTTP/1.1 411 Length Required\r\n\r\n"
					} else {
						// Find the start of the body
						headerEnd := strings.Index(req, "\r\n\r\n")
						if headerEnd == -1 {
							res = "HTTP/1.1 400 Bad Request\r\n\r\n"
						} else {
							body := buf[headerEnd+4 : headerEnd+4+contentLength] // Extract body based on Content-Length
							if err := os.WriteFile(*dir+filename, body, 0644); err == nil {
								fmt.Println("File written:", *dir+filename)
								res = "HTTP/1.1 201 Created\r\nContent-Length: 0\r\n\r\n"
							} else {
								fmt.Println("Failed to write file:", err)
								res = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
							}
						}
					}
				}
			} else {
				res = "HTTP/1.1 404 Not found\r\n\r\n"
			}
			fmt.Println(res)
			conn.Write([]byte(res))
			if err != nil {
				fmt.Println("Error accepting connection: ", err)
				os.Exit(1)
			}
		}()
	}
}