package main

import (
	"flag"
	"fmt"
	"github.com/mildred/simpleipc"
	"log"
	"net"
)

func main() {
	var sockPath string
	flag.StringVar(&sockPath, "sock", "sock", "Socker file path")
	flag.Parse()

	err := run(sockPath)
	if err != nil {
		log.Fatal(err)
	}
}

func run(sockPath string) error {
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		return err
	}

	for {
		cnx, err := l.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go (func() {
			err := handleClient(cnx.(*net.UnixConn))
			if err != nil {
				log.Print(err)
			}
		})()
	}

	return nil
}

func handleClient(cnx *net.UnixConn) error {
	var h simpleipc.Header
	log.Printf("Receive client %v", cnx)
	err := h.Read(cnx, nil)
	if err != nil {
		return err
	}
	log.Printf("Received header %#v", h)
	if len(h.Files) > 0 {
		fmt.Fprintf(h.Files[0], "Hello world\n")
	}
	return nil
}
