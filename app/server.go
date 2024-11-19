package main
import (
	"fmt"
	"net"
	"os"
	"strings"
)

const CRLF = "\r\n"

func handleConnection(conn net.Conn, dir string) {
	defer conn.Close()
	buf := make([]byte, 1024)
	conn.Read(buf)
	fmt.Print(string(buf))
	req := string(buf)
	lines := strings.Split(req, CRLF)
	path := strings.Split(lines[0], " ")[1]
	method := strings.Split(lines[0], " ")[0]
	fmt.Println(path)
	var pathUA string
	headers := strings.Split(req, "\r\n")
	for _, header := range headers {
		if strings.HasPrefix(header, "User-Agent:") {
			uaParts := strings.SplitN(header, ":", 2)
			pathUA = strings.TrimSpace(uaParts[1])
			break
		}
	}
	response := ""
	if strings.HasPrefix(req, "GET / HTTP") {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.Contains(req, "/echo/") {
		echo := strings.TrimPrefix(path, "/echo/")
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
	} else if strings.Contains(req, "/user-agent") && pathUA != "" {
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(pathUA), pathUA)
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
			file := []byte(strings.Trim(lines[6], "\x00"))
			if err := os.WriteFile(*dir+filename, file, 0644); err == nil {
				fmt.Println("wrote file")
				res = "HTTP/1.1 201 OK\r\n\r\n"
			} else {
				res = "HTTP/1.1 404 Not found\r\n\r\n"
			}
		}
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	conn.Write([]byte(response))
}
func main() {
	fmt.Println("Logs from your program will appear here!")
	dir := flag.String("directory", "", "enter a directory")
	flag.Parse()
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
		go handleConnection(conn, *dir)
	}
}