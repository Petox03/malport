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

// --- MODELOS DEL PROTOCOLO ---
type Request struct {
	Method   string
	Path     string
	HTTPV    string
	Headers  map[string]string
	RawLines []string
	Body     []byte
}

type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

// --- CAPA DE TRANSPORTE ---
func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Net Error: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Printf("Listening on: %s (Network: %s)\n", ln.Addr().String(), ln.Addr().Network())
	acceptLoop(ln)
}

func acceptLoop(ln net.Listener) {
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

	// Parsear entrada
	req, err := parseRequest(reader)
	if err != nil {
		// Early return. El mensaje exacto ya se imprimió en parseRequest
		return
	}

	// Procesar lógica
	res := handleRequest(req)

	// Escribir salida
	if err := writeResponse(c, res); err != nil {
		fmt.Printf("Error writing response: %v\n", err)
	}
}

// --- CAPA DE PROTOCOLO HTTP ---
func parseRequest(reader *bufio.Reader) (*Request, error) {
	req := &Request{
		Headers: make(map[string]string),
	}

	// Leer Request Line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Error reading request line: %v\n", err)
		}
		return nil, err
	}
	req.RawLines = append(req.RawLines, requestLine)

	parts := strings.Fields(requestLine)
	if len(parts) < 3 {
		err := fmt.Errorf("malformed request line: %q", strings.TrimSpace(requestLine))
		fmt.Println(err)
		return nil, err
	}

	req.Method = parts[0]
	req.Path = parts[1]
	req.HTTPV = parts[2]

	// Leer Headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading headers: %v\n", err)
			return nil, err
		}
		req.RawLines = append(req.RawLines, line)

		if line == "\r\n" {
			break
		}

		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			key := strings.ToLower(strings.TrimSpace(headerParts[0]))
			val := strings.TrimSpace(headerParts[1])
			req.Headers[key] = val
		}
	}

	// Leer Body si existe Content-Length
	if clStr, ok := req.Headers["content-length"]; ok {
		contentLength, err := strconv.Atoi(clStr)
		if err != nil || contentLength < 0 {
			err := fmt.Errorf("invalid content-length: %q", clStr)
			fmt.Println(err)
			return nil, err
		}
		if contentLength > 0 {
			req.Body = make([]byte, contentLength)
			if _, err := io.ReadFull(reader, req.Body); err != nil {
				fmt.Printf("Error reading body: %v\n", err)
				return nil, err
			}
		}
	}

	return req, nil
}

func writeResponse(w io.Writer, res *Response) error {
	if res.Headers == nil {
		res.Headers = make(map[string]string)
	}
	res.Headers["Content-Length"] = strconv.Itoa(len(res.Body))

	var sb strings.Builder
	fmt.Fprintf(&sb, "HTTP/1.1 %d %s\r\n", res.StatusCode, res.StatusText)

	for k, v := range res.Headers {
		fmt.Fprintf(&sb, "%s: %s\r\n", k, v)
	}
	sb.WriteString("\r\n")

	if _, err := w.Write([]byte(sb.String())); err != nil {
		return err
	}
	_, err := w.Write(res.Body)
	return err
}

// --- CAPA DE NEGOCIO / APLICACIÓN ---
func handleRequest(req *Request) *Response {
	// Imprimir cada línea cruda (Request Line + Headers + \r\n) igual que el original
	for _, line := range req.RawLines {
		fmt.Print(line)
	}

	// Imprimir el body solo si existe, con el formato exacto del original
	if len(req.Body) > 0 {
		fmt.Printf("Method: %v\n", string(req.Method))
		fmt.Printf("Path: %v\n", string(req.Path))
		fmt.Printf("HTTP Version: %v\n", string(req.HTTPV))
		fmt.Printf("Received body: %s\n\n", string(req.Body))
	}

	return &Response{
		StatusCode: 200,
		StatusText: "OK",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"status": "ok", "received": true}`),
	}
}
