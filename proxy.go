package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("SOCKS5 server listening on 1080..")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go serveClient(conn)

		defer conn.Close()
	}
}

func serveClient(conn net.Conn) {
	buf := make([]byte, 256)

	//read the Socks version from the client
	_, err := conn.Read(buf)
	fmt.Printf("Bytes read: %x\n", buf)
	if err != nil {
		log.Println("error reading bytes for version")
	}

	if buf[0] == 0x05 {
		if buf[2] != 0x00 {
			log.Println("Client needs authentication..")
			return
		}

		log.Println(conn.RemoteAddr().String())

		// Send server greeting
		_, err = conn.Write([]byte{0x05, 0x00})
		if err != nil {
			log.Println("Failed to send server greeting:", err)
			return
		} else {
			fmt.Println("sent")
		}

		_, err = conn.Read(buf)
		fmt.Printf("Bytes read 2 : %x\n", buf)

		var target string
		//0x01: IPv4 address, followed by 4 bytes IP
		if buf[3] == 0x01 {
			target := net.IP(buf[4:8]).String()
			fmt.Printf("Address target : %s\n", target)
			_, err = conn.Write(buf)
		}

		port := binary.BigEndian.Uint16(buf[8:10])
		fmt.Print("Port: ", port)

		server, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", target, port), 10*time.Second)
		if err != nil {
			fmt.Print("Error: ", err)
			return
		}

		defer server.Close()

		// Respond with a success message to the client.
		response := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		conn.Write(response)
		fmt.Print("Done here")
		// Start bi-directional data transfer.
		go io.Copy(server, conn)
		io.Copy(conn, server)

	} else {
		fmt.Println("Unsupported SOCKS version:", buf[0])
		return
	}
	// 05 01 00 01 68 15 52 fa 01 bb 00
}
