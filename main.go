package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-routeros/routeros"
	"github.com/joho/godotenv"
)

func main() {
	command := flag.String("c", "", "Mikrotik command to execute")
	flag.Parse()

	if *command == "" {
		fmt.Println("Usage: mt -c '<command> [args...]'")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  mt -c '/system/resource/print'")
		fmt.Println("  mt -c '/interface/print'")
		fmt.Println("  mt -c '/ip/service/set =.id=*0 =address=10.11.13.0/24'")
		fmt.Println()
		fmt.Println("Filtering with 'where' (use ? prefix):")
		fmt.Println("  mt -c '/ip/service/print ?name=api'              # where name = api")
		fmt.Println("  mt -c '/interface/print ?type=ether'             # where type = ether")
		fmt.Println("  mt -c '/ip/address/print ?interface=bridge1'     # where interface = bridge1")
		fmt.Println("  mt -c '/interface/print ?running=true'           # where running = true")
		os.Exit(1)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("MT_HOST")
	user := os.Getenv("MT_USER")
	password := os.Getenv("MT_PASSWORD")
	port := os.Getenv("MT_PORT")
	useTLS := os.Getenv("MT_USE_TLS")

	if host == "" || user == "" || password == "" {
		log.Fatal("MT_HOST, MT_USER, and MT_PASSWORD must be set in .env")
	}

	if port == "" {
		if useTLS == "true" {
			port = "8729"
		} else {
			port = "8728"
		}
	}

	address := fmt.Sprintf("%s:%s", host, port)

	var client *routeros.Client
	var err error

	if useTLS == "true" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		client, err = routeros.DialTLS(address, user, password, tlsConfig)
	} else {
		client, err = routeros.Dial(address, user, password)
	}

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Split command into words
	args := strings.Fields(*command)

	reply, err := client.Run(args...)
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	for _, re := range reply.Re {
		for _, pair := range re.List {
			fmt.Printf("%s: %s\n", pair.Key, pair.Value)
		}
		fmt.Println()
	}
}
