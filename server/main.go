package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	port := flag.Int("-port", 9090, "TCP listen port")
	chanList := flag.String("-channels", "golang;c++", "channel list")
	flag.Parse()
	chanMgr := NewChannelManager()
	chs := strings.Split(*chanList, ";")
	for _, ch := range chs {
		chanMgr.CreateChannel(ch)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("Listen failure error=[%v]", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Accept eror=[%v]", err)
		}
		client := NewClient(conn, chanMgr)

		go client.MainLoop()
	}
}
