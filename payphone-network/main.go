package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

func main() {
	listenPort := flag.String("listen", "9000", "Port to listen on for incoming connections")
	targetAddr := flag.String("connect", "", "Target peer address")
	username := flag.String("user", "node", "Identifier displayed to the peer")
	flag.Parse()

	var conn net.Conn
	var err error

	if *targetAddr != "" {
		fmt.Printf("[*] Dialing peer at %s...\n", *targetAddr)
		conn, err = net.Dial("tcp", *targetAddr)
		if err != nil {
			log.Fatalf("[-] Connection failed: %v", err)
		}
		fmt.Printf("[+] Connected to %s\n", *targetAddr)
	} else {
		ln, err := net.Listen("tcp", ":"+*listenPort)
		if err != nil {
			log.Fatalf("[-] Failed to bind port %s: %v", *listenPort, err)
		}
		defer ln.Close()

		fmt.Printf("[*] Listening on port %s. Waiting for peer...\n", *listenPort)
		conn, err = ln.Accept()
		if err != nil {
			log.Fatalf("[-] Accept error: %v", err)
		}
		fmt.Printf("[+] Peer connected from %s\n", conn.RemoteAddr().String())
	}

	defer conn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		reader := bufio.NewReader((conn))
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("\n[-] Peer severed connection.")
				} else {
					fmt.Printf("\n[-] Read error: %v\n", err)
				}
				return
			}
			fmt.Print(line)
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "" {
				continue
			}
			if text == "/exit" || text == "/quit" {
				fmt.Println("[*] Closing link...")
				conn.Close()
				return
			}

			payload := fmt.Sprintf("[%s]: %s\n", *username, text)
			_, err := conn.Write([]byte(payload))
			if err != nil {
				fmt.Printf("[-] Failed to send: %v\n", err)
				return
			}
		}
	}()

	wg.Wait()
	fmt.Println("[*] Session terminated")
}
