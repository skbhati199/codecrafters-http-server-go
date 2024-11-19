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
func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()
	buf := make([]byte, 1024)
	conn.Read(buf)
	fmt.Print(string(buf))
	req := string(buf)
	path := strings.Split(req, "\r\n")[0]
	path = strings.TrimSpace(path)
	path = strings.Split(path, " ")[1]
	var pathUA string
	headers := strings.Split(req, "\r\n")
	for _, header := range headers {
		if strings.HasPrefix(header, "User-Agent:") {
			uaParts := strings.SplitN(header, ":", 2)
			pathUA = strings.TrimSpace(uaParts[1])
			break
		}
	}

	reader := bufio.NewReader(conn)

	// Read the request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request:", err.Error())
		return
	}

	method, path := extractMethodAndPath(requestLine)

	// Handle the request
	switch method {
	case "GET":
		handleGet(conn, directory, path)
	case "POST":
		contentLength, _ := strconv.Atoi(headers["content-length"])
		handlePost(conn, directory, path, reader, contentLength)
	default: {
		response := ""
		if strings.HasPrefix(req, "GET / HTTP") {
			response = "HTTP/1.1 200 OK\r\n\r\n"
		} else if strings.Contains(req, "/echo/") {
			echo := strings.TrimPrefix(path, "/echo/")
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
		} else if strings.Contains(req, "/user-agent") && pathUA != "" {
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(pathUA), pathUA)
		} else if strings.Contains(req, "/files/") {
			dir := os.Args[2]
			fileName := strings.TrimPrefix(path, "/files/")
			fmt.Print(fileName)
			data, err := os.ReadFile(dir + fileName)
			if err != nil {
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			} else {
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
			}
		} else {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		}
		conn.Write([]byte(response))
		}
	}

}
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
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, *directory)
	}
}

func extractMethodAndPath(requestLine string) (string, string) {
	parts := strings.Split(requestLine, " ")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], parts[1]
}