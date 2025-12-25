package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-routeros/routeros"
	"github.com/joho/godotenv"
)

func main() {
	host := flag.String("h", "", "Mikrotik host address")
	user := flag.String("u", "", "Username")
	password := flag.String("p", "", "Password")
	port := flag.String("P", "", "API port (default: 8728, or 8729 with TLS)")
	useTLS := flag.Bool("tls", false, "Use TLS connection")
	command := flag.String("c", "", "Command to execute")
	flag.Parse()

	if *command == "" {
		fmt.Println("Usage: mt [-h host] [-u user] [-p pass] -c '<command>'")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -h host    Mikrotik host (or MT_HOST env)")
		fmt.Println("  -u user    Username (or MT_USER env)")
		fmt.Println("  -p pass    Password (or MT_PASSWORD env)")
		fmt.Println("  -P port    API port (default: 8728, or 8729 with -tls)")
		fmt.Println("  -tls       Use TLS connection")
		fmt.Println("  -c cmd     Command to execute")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  mt -c '/system/resource/print'")
		fmt.Println("  mt -h 10.11.13.63 -u admin -p secret -c '/interface/print'")
		fmt.Println("  mt -c '/ip/service/print ?name=api'")
		fmt.Println()
		fmt.Println("Filtering (use ? prefix):")
		fmt.Println("  mt -c '/interface/print ?type=ether'")
		fmt.Println("  mt -c '/interface/print ?running=true'")
		os.Exit(1)
	}

	// Load .env if present
	_ = godotenv.Load()

	// Use flags if provided, otherwise fall back to env
	if *host == "" {
		*host = os.Getenv("MT_HOST")
	}
	if *user == "" {
		*user = os.Getenv("MT_USER")
	}
	if *password == "" {
		*password = os.Getenv("MT_PASSWORD")
	}
	if *port == "" {
		*port = os.Getenv("MT_PORT")
	}
	if !*useTLS && os.Getenv("MT_USE_TLS") == "true" {
		*useTLS = true
	}

	if *host == "" || *user == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: host, user, and password required (via flags or .env)")
		os.Exit(1)
	}

	if *port == "" {
		if *useTLS {
			*port = "8729"
		} else {
			*port = "8728"
		}
	}

	address := fmt.Sprintf("%s:%s", *host, *port)

	var client *routeros.Client
	var err error

	if *useTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		client, err = routeros.DialTLS(address, *user, *password, tlsConfig)
	} else {
		client, err = routeros.Dial(address, *user, *password)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to connect to %s: %v\n", address, err)
		os.Exit(1)
	}
	defer client.Close()

	args := strings.Fields(*command)

	reply, err := client.Run(args...)
	if err != nil {
		if strings.Contains(err.Error(), "!empty") {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(reply.Re) == 0 {
		os.Exit(0)
	}

	for _, re := range reply.Re {
		for _, pair := range re.List {
			fmt.Printf("%s: %s\n", pair.Key, pair.Value)
		}
		fmt.Println()
	}
}
