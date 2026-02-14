package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	updAdrr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("error")
		return
	}
	udpConn, err := net.DialUDP(updAdrr.Network(), nil, updAdrr)
	if err != nil {
		log.Fatal("error")
		return
	}
	defer udpConn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		data, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, err = udpConn.Write([]byte(data))
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
