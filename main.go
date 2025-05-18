package main

import (
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	curr "github.com/vladimirvivien/learning-go/ch11/curr0"
)

var currencies = curr.Load("./data.csv")

func main() {
	ln, err := net.Listen("tcp", ":4040")
	if err != nil {
		fmt.Println("error starting the server", err)
		return
	}
	slog.Info("Server up and running")
	defer ln.Close()
	// connection loop
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			conn.Close()
			continue
		}
		go handleConnection(conn)
	}
}

// handle client connection
func handleConnection(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
	slog.Info("New connection", "client", remoteAddr)
	// loop to stay connected with client
	for {
		cmdLine := make([]byte, (1024 * 4))
		n, err := conn.Read(cmdLine)
		if n == 0 || err != nil {
			slog.Info("Connection Closed", "client", remoteAddr, "reason", "read error", "details", err)
			return
		}
		//logiing raw received data
		slog.Debug("Received data", "client", remoteAddr, "bytes", n, "content", string(cmdLine[:n]))
		cmd, param := parseCommand(string(cmdLine[0:n]))
		if cmd == "" {
			slog.Warn("Invalid command format", "client", remoteAddr, "input", string(cmdLine[:n]))
			continue
		}
		slog.Info("Command received",
			"client", remoteAddr,
			"command", cmd,
			"parameter", param)
		// execute command
		switch strings.ToUpper(cmd) {
		case "GET":
			result := curr.Find(currencies, param)
			slog.Info("Processing GET request",
				"client", remoteAddr,
				"parameter", param,
				"results", len(result))
			// stream result to client
			for _, cur := range result {
				_, err := fmt.Fprintf(
					conn,
					"%s %s %s %s\n",
					cur.Name, cur.Code,
					cur.Number, cur.Country,
				)
				if err != nil {
					slog.Warn("Write error",
						"client", remoteAddr,
						"error", err)
					return
				}
				// reset deadline while writing,
				// closes conn if client is gone
				conn.SetWriteDeadline(
					time.Now().Add(time.Second * 5))
			}
			// reset read deadline for next read
			conn.SetReadDeadline(
				time.Now().Add(time.Second * 300))
			slog.Info("GET request completed",
				"client", remoteAddr,
				"parameter", param)
		default:
			slog.Warn("Invalid command", "client", remoteAddr, "command", cmd)
			conn.Write([]byte("Invalid command\n"))
		}
	}
}
func parseCommand(cmdLine string) (cmd, param string) {
	parts := strings.Split(cmdLine, " ")
	if len(parts) != 2 {
		return "", ""
	}
	cmd = strings.TrimSpace(parts[0])
	param = strings.TrimSpace(parts[1])
	return
}
