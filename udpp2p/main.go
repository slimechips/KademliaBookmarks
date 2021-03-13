package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const port = 1053

func main() {
	// net interface
	ip := getIp()
	nodeName := getNodeName(ip)
	fmt.Println(nodeName)
	// listen to incoming udp packets
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		go serve(pc, addr, buf[:n])
	}

}

func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	// 0 - 1: ID
	// 2: QR(1): Opcode(4)
	buf[2] |= 0x80 // Set QR bit

	pc.WriteTo(buf, addr)
	fmt.Println("Received")

}

func getIp() net.IP {
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		if i.Name != "eth0" {
			continue
		}
		addrs, _ := i.Addrs()
		var ip net.IP
		switch v := addrs[0].(type) {
		case *net.IPAddr:
			ip = v.IP
		case *net.IPNet:
			ip = v.IP
		}
		return ip
	}
	return nil
}

func getNodeName(ip net.IP) int {
	ipArr := strings.Split(ip.String(), ".")
	nodeName, _ := strconv.Atoi(ipArr[3])
	return nodeName
}
