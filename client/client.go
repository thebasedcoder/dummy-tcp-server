package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"
)

var host, port = "127.0.0.1", "4040"
var addr = net.JoinHostPort(host, port)

const prompt = "curr"
const buffLen = 1024

func main() {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		slog.Info("Error sstablishing connection", "remote address", addr)
		return
	}
	defer conn.Close()

	var cmd, param string
	for {
		fmt.Print(prompt, "> ")
		_, err = fmt.Scanf("%s %s", &cmd, &param)
		if err != nil {
			fmt.Println("Usage: GET <search string or *>")
			continue
		}
		cmdLine := fmt.Sprintf("%s %s", cmd, param)
		if n, err := conn.Write([]byte(cmdLine)); n == 0 || err != nil {
			slog.Warn("Error sending the command to server", "error", err)
			return
		}
		conn.SetReadDeadline(time.Now().Add(time.Second * 5))

		for {
			buff := make([]byte, buffLen)
			n, err := conn.Read(buff)
			if err != nil {
				if err == io.EOF {
					slog.Info("Server closed connection gracefully")
					break
				}
				slog.Error("Read failed", "error", err, "partial_data", string(buff[:n]))

				break
			}
			response := string(buff[:n])
			slog.Info("Server response",
				"bytes", n,
				"content", response,
			)
			fmt.Print(response)
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 700))
		}
	}

}
