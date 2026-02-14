package main

import (
	"fmt"
	"httpProtocols/internal/request"
	"log"
	"net"
)

func main() {
	listner, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("something happened...")
		return
	}
	defer listner.Close()
	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Fatal("can't establish connection")
			return
		}
		fmt.Println("connection has been accepted.")
		req, err := request.RequestFromReader(conn)
		if err != nil || req == nil {
			fmt.Println("an error occurred\n ", err.Error())
			continue
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n- Target: %v\n- Version: %v\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
	}

}
