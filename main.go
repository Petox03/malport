package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Net Error: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Printf("Listening on: %s (Network: %s)\n", ln.Addr().String(), ln.Addr().Network())

	for {
		// Esperar a que alguien se conecte
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		// Manejar cada cliente en una goroutine independiente
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()

	reader := bufio.NewReader(c)
	contentLength := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading headers: %v\n", err)
			return
		}
		fmt.Print(line)

		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			value := strings.TrimSpace(strings.Split(line, ":")[1])
			contentLength, _ = strconv.Atoi(value)
		}

		if line == "\r\n" {
			break
		}
	}

	if contentLength > 0 {
		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			fmt.Printf("Error reading body: %v\n", err)
			return
		}
		fmt.Printf("Received body: %s\n", string(body))
	}

	responseBody := `{"status": "ok", "received": true}`
	httpResponse := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s",
		len(responseBody),
		responseBody,
	)

	_, err := c.Write([]byte(httpResponse))
	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
	}
}
