package main
import (
	"fmt"
	"net"
	"flag"
	"os"
	"path"
	"strings"
)

const CRLF = "\r\n"

func handleConnection(conn net.Conn) {
	defer conn.Close()
	flag.Parse()
	buf := make([]byte, 1024)
	conn.Read(buf)
	fmt.Print(string(buf))
	r := strings.Split(string(buf), "\r\n")
	m := strings.Split(r[0], " ")[0]
	p := strings.Split(r[0], " ")[1]
	req := string(buf)
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
		echo := strings.TrimPrefix(p, "/echo/")
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
	} else if strings.Contains(req, "/user-agent") && pathUA != "" {
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(pathUA), pathUA)
	} else if m == "GET" && p[0:7] == "/files/" {
		dir := os.Args[2]
		content, err := os.ReadFile(path.Join(dir, p[7:]))
		if err != nil {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), string(content))
		}
	} else if m == "POST" && p[0:7] == "/files/" {
		content := strings.Trim(r[len(r)-1], "\x00")
		directory := flag.String("directory", "", "the directory to serve files from")
	flag.Parse()

	if *directory == "" {
		fmt.Println("Please provide a directory using the --directory flag")
		os.Exit(1)
	}
		_ = os.WriteFile(path.Join(directory, p[7:]), []byte(content), 0644)
		response = "HTTP/1.1 201 Created\r\n\r\n"
	}  else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	conn.Write([]byte(response))
}
func main() {
	fmt.Println("Logs from your program will appear here!")
	
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
		go handleConnection(conn)
	}
}