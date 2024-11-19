package main
import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)
var filesDir string
func main() {
	flag.StringVar(&filesDir, "directory", "./files", "specify a directory where files uploaded to server are saved")
	flag.Parse()
	fmt.Println("Saving files to: ", filesDir)
	fmt.Println("Got args: ", os.Args)
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Println("Listening")
	for {
		conn, err := l.Accept()
		fmt.Println("Accepted")
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn)
		fmt.Println("Handling started")
	}
}
func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	request := string(buffer[:n])
	resp := dispatch(parseRequest(request))
	conn.Write([]byte(resp))
	fmt.Println("Handled")
}
func parseRequest(request string) (method string, path string, headers map[string]string, body string) {
	fmt.Println("~~~ request ~~~\n", request)
	requestParts := strings.Split(request, "\r\n")
	requestLine := requestParts[0]
	lineParts := strings.Split(requestLine, " ")
	method = lineParts[0]
	path = lineParts[1]
	body = requestParts[len(requestParts)-1]
	headers = make(map[string]string)
	for i := 1; i < len(requestParts)-1; i++ {
		key, value, _ := strings.Cut(requestParts[i], ":")
		headers[strings.ToLower(strings.Trim(key, " "))] = strings.Trim(value, " ")
	}
	return // named result parameters
}
func dispatch(method string, urlPath string, headers map[string]string, body string) string {
	if method != "GET" {
		if method == "POST" && strings.HasPrefix(urlPath, "/files/") {
			fname := urlPath[len("/files/"):]
			f, err := os.Create(path.Join(filesDir, fname))
			if err != nil {
				return "HTTP/1.1 422 Unprocessable Entity\r\n\r\n"
			}
			defer f.Close()
			f.WriteString(body)
			return "HTTP/1.1 201 Created\r\n\r\n"
		} else {
			return "HTTP/1.1 405 Method Not Allowed\r\n\r\n"
		}
	}
	if urlPath == "/" || urlPath == "/index.html" {
		return "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(urlPath, "/echo/") {
		text := urlPath[len("/echo/"):]
		return contentResponse(text, "text/plain", headers)
	} else if urlPath == "/user-agent" {
		agent := headers["user-agent"]
		return contentResponse(agent, "text/plain", headers)
	} else if strings.HasPrefix(urlPath, "/files/") {
		fname := urlPath[len("/files/"):]
		dat, err := os.ReadFile(path.Join(filesDir, fname))
		if err != nil {
			return "HTTP/1.1 404 Not Found\r\n\r\n"
		}
		return contentResponse(string(dat), "application/octet-stream", headers)
	} else {
		return "HTTP/1.1 404 Not Found\r\n\r\n"
	}
}
func contentResponse(content string, contentType string, requestHeaders map[string]string) string {
	values, ok := requestHeaders["accept-encoding"]
	compression := ""
	if ok {
		for _, elem := range strings.Split(values, ",") {
			if strings.TrimSpace(elem) == "gzip" {
				var buffer bytes.Buffer
				w := gzip.NewWriter(&buffer)
				w.Write([]byte(content))
				w.Close()
				content = buffer.String()
				compression = "Content-Encoding: gzip\r\n"
				break
			}
		}
	}
	return "HTTP/1.1 200 OK\r\n" +
		compression +
		"Content-Type: " + contentType + "\r\n" +
		"Content-Length: " + fmt.Sprint(len(content)) +
		"\r\n" +
		"\r\n" + content
}