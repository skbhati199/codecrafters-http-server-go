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
			n, err = conn.Read(buf)
			if err != nil {
				fmt.Println("Error accepting connection: ", err)
			}
			body := strconv.Quote(string(buf[:n]))
			req, _ := parseRequest(body)
			lines := strings.Split(req, CRLF)
			path := strings.Split(lines[0], " ")[1]
			method := strings.Split(lines[0], " ")[0]
			fmt.Println(path)
			var res string
			if path == "/" {
				res = "HTTP/1.1 200 OK\r\n\r\n"
			} else if strings.HasPrefix(path, "/echo/") {
				echoOut := req.path[strings.Index(req.path, "/echo/")+len("/echo/"):]
				// res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(msg), msg)
				var contEncoding = ""
				if req.encoding == "gzip" {
					fmt.Println("in encoding")
					contEncoding = fmt.Sprintf("Content-Encoding: %s\r\n", req.encoding)
				}
				fmt.Println(contEncoding)
				res := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n%sContent-Length: %d\r\n\r\n%s", contEncoding, len(echoOut), echoOut)
				writeResponse(conn, res)
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
						res = "HTTP/1.1 404 Not Found\r\n\r\n"
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
				res = "HTTP/1.1 404 Not Found\r\n\r\n"
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

func writeResponse(conn net.Conn, res string) {
	_, err := conn.Write([]byte(res))
	if err != nil {
		fmt.Println("failed to write to connection")
		return
	}
}

func parseRequest(request string) (*HTTPRequest, error) {
	strs := strings.Split(request, "\\r\\n")
	req := HTTPRequest{}
	for _, item := range strs {
		if strings.Contains(item, "GET") || strings.Contains(item, "POST") {
			headerParts := strings.Fields(item)
			// set http verb
			req.verb = strings.Trim(headerParts[0], "\"")
			// set route
			req.path = headerParts[1]
			// set http version
			req.httpVersion = headerParts[2]
		}
		if strings.Contains(item, "Host: ") {
			req.host = item[strings.Index("Host: ", item)+len("Host: "):]
		}
		if strings.Contains(item, "User-Agent: ") {
			req.userAgent = item[strings.Index("User-Agent: ", item)+len("User-Agent: "):]
		}
		if strings.Contains(item, "Accept-Encoding: ") {
			req.encoding = item[strings.Index("Accept-Encoding: ", item)+len("Accept-Encoding: "):]
			req.encoding = strings.Trim(req.encoding, " ")
		}
	}
	if req.verb == "POST" {
		req.body = strings.TrimSuffix(strs[len(strs)-1], "\"")
	}
	return &req, nil
}