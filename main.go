package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	curr "github.com/vladimirvivien/learning-go/ch11/curr0"
)

var currencies = curr.Load("./data.csv")

func main() {
	ln, _ := net.Listen("tcp", ":4040")
	defer ln.Close()
	// connection loop
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			conn.Close()
			continue
		}
		fmt.Printf("connected to %s", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

// handle client connection
func handleConnection(conn net.Conn) {
	defer conn.Close()
	// loop to stay connected with client
	for {
		cmdLine := make([]byte, (1024 * 4))
		n, err := conn.Read(cmdLine)
		if n == 0 || err != nil {
			return
		}
		cmd, param := parseCommand(string(cmdLine[0:n]))
		if cmd == "" {
			continue
		}
		// execute command
		switch strings.ToUpper(cmd) {
		case "GET":
			result := curr.Find(currencies, param)
			// stream result to client
			for _, cur := range result {
				_, err := fmt.Fprintf(
					conn,
					"%s %s %s %s\n",
					cur.Name, cur.Code,
					cur.Number, cur.Country,
				)
				if err != nil {
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
		default:
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
