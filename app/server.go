package main
import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)
// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit
type httpRequestDetails struct {
	method         []byte
	path           []byte
	host           []byte
	userAgent      []byte
	requestBody    []byte
	acceptEncoding []byte
}
type httpRequest struct {
	statusLine  []byte
	headers     [][]byte
	requestBody []byte
}
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err1 := l.Accept()
		if err1 != nil {
			fmt.Println("Error accepting connection: ", err1.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	readBuf := make([]byte, 256)
	readByte, readErr := conn.Read(readBuf)
	if readErr != nil {
		fmt.Println("Couldn't read data:", readErr)
	}
	defer conn.Close()
	rq := requestResult(readBuf, readByte)
	if string(rq.path) == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.Contains(string(rq.path), "echo") {
		msg := splitPath(string(rq.path))[2]
		if string(rq.acceptEncoding) == "gzip" {
		if acceptedEncoding(rq) {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nContent-Encoding: gzip\r\n\r\n%s", len(msg), msg)))
		} else {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(msg), msg)))
		}
	}
	} else if strings.Contains(string(rq.path), "files") && string(rq.method) == "GET" {
		fileName := splitPath(string(rq.path))[2]
		fileContent, fileRequestErr := handleFileReadRequest(fileName)
		if fileRequestErr != nil {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(fileContent), fileContent)))
	} else if strings.Contains(string(rq.path), "user-agent") {
		userAgent := string(rq.userAgent)
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s/1.2.3", len(userAgent), userAgent)))
	} else if string(rq.method) == "POST" && strings.Contains(string(rq.path), "files") {
		fileName := splitPath(string(rq.path))[2]
		fileRequestErr := handleFileWriteRequest(fileName, rq.requestBody)
		if fileRequestErr != nil {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 201 Created\r\n\r\n")))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	splitRequest(readBuf, readByte)
}

func splitRequest(readBuf []byte, readByte int) httpRequest {
	splittedReq := bytes.Split(readBuf[:readByte], []byte("\r\n"))
	statusLine := splittedReq[0]
	headers := splittedReq[1 : len(splittedReq)-2]
	requestBody := splittedReq[len(splittedReq)-1]
	return httpRequest{
		statusLine:  statusLine,
		headers:     headers,
		requestBody: requestBody,
	}
}
func mapHeaders(headers [][]byte) map[string]string {
	res := make(map[string]string)
	for i := range headers {
		line := strings.Split(string(headers[i]), ": ")
		res[line[0]] = line[1]
	}
	return res
}


func requestResult(readBuf []byte, readByte int) httpRequestDetails {
	sreq := splitRequest(readBuf, readByte)
	statusLine := bytes.Split(sreq.statusLine, []byte(" "))
	headerMap := mapHeaders(sreq.headers)
	requestBody := sreq.requestBody
	return httpRequestDetails{
		method:         statusLine[0],
		path:           statusLine[1],
		host:           []byte(headerMap["Host"]),
		userAgent:      []byte(headerMap["User-Agent"]),
		requestBody:    requestBody,
		acceptEncoding: []byte(headerMap["Accept-Encoding"]),
	}
}


func splitPath(s string) []string {
	return strings.Split(s, "/")
}


func handleFileReadRequest(fileName string) ([]byte, error) {
	arguments := os.Args[1:]
	var directory string
	if len(arguments) <= 0 {
		fmt.Println(fmt.Errorf("please provide some arguments"))
		return []byte(""), fmt.Errorf("please provide some arguments")
	} else if arguments[len(arguments)-1] == "--directory" {
		fmt.Println(fmt.Errorf("argument --directory specified, but value didn't provided"))
		return []byte(""), fmt.Errorf("argument --directory specified, but value didn't provided")
	}
	exist := false
	for i := range arguments {
		if arguments[i] == "--directory" && arguments[len(arguments)-1] != "--directory" {
			directory = arguments[i+1]
			exist = true
		}
	}
	if !exist {
		fmt.Println(fmt.Errorf("no --directory argument passwd. please provide --directory argument and restart the program"))
		return []byte(""), fmt.Errorf("no --directory argument passwd. please provide --directory argument and restart the program")
	}
	fileContent, fileContentErr := os.ReadFile(directory + "/" + fileName)
	if fileContentErr != nil {
		fmt.Println(fileContentErr)
		return []byte(""), fileContentErr
	}
	return fileContent, nil
}
func handleFileWriteRequest(fileName string, fileContent []byte) error {
	arguments := os.Args[1:]
	var directory string
	if len(arguments) <= 0 {
		fmt.Println(fmt.Errorf("please provide some arguments"))
		return fmt.Errorf("please provide some arguments")
	} else if arguments[len(arguments)-1] == "--directory" {
		fmt.Println(fmt.Errorf("argument --directory specified, but value didn't provided"))
		return fmt.Errorf("argument --directory specified, but value didn't provided")
	}
	exist := false
	for i := range arguments {
		if arguments[i] == "--directory" && arguments[len(arguments)-1] != "--directory" {
			directory = arguments[i+1]
			exist = true
		}
	}
	if !exist {
		fmt.Println(fmt.Errorf("no --directory argument passwd. please provide --directory argument and restart the program"))
		return fmt.Errorf("no --directory argument passwd. please provide --directory argument and restart the program")
	}
	fileContentErr := os.WriteFile(directory+"/"+fileName, fileContent, 0644)
	if fileContentErr != nil {
		fmt.Println(fileContentErr)
		return fileContentErr
	}
	return nil
}
func acceptedEncoding(rq httpRequestDetails) bool {
	var res bool
	acceptedEncodings := strings.Split(string(rq.acceptEncoding), ", ")
	for _, encoding := range acceptedEncodings {
		if encoding == "gzip" {
			res = true
		}
	}
	return res
}