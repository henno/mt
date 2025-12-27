package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-routeros/routeros"
	"github.com/joho/godotenv"
)

func main() {
	host := flag.String("h", "", "Mikrotik host address")
	user := flag.String("u", "", "Username")
	password := flag.String("p", "", "Password")
	port := flag.String("P", "", "Port (default: 8728 API, 8729 TLS, 22 SSH)")
	useTLS := flag.Bool("tls", false, "Use TLS connection")
	useSSH := flag.Bool("ssh", false, "Use SSH connection (CLI mode)")
	command := flag.String("c", "", "Command to execute")
	flag.Parse()

	if *command == "" {
		fmt.Println("Usage: mt [-h host] [-u user] [-p pass] -c '<command>'")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -h host    Mikrotik host (or MT_HOST env)")
		fmt.Println("  -u user    Username (or MT_USER env)")
		fmt.Println("  -p pass    Password (or MT_PASSWORD env)")
		fmt.Println("  -P port    Port (default: 8728 API, 8729 TLS, 22 SSH)")
		fmt.Println("  -tls       Use TLS connection (API mode)")
		fmt.Println("  -ssh       Use SSH connection (CLI mode)")
		fmt.Println("  -c cmd     Command to execute")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  mt -c '/system/resource/print'")
		fmt.Println("  mt -ssh -c '/system resource print'")
		fmt.Println("  mt -h 10.11.13.63 -u admin -p secret -c '/interface/print'")
		fmt.Println()
		fmt.Println("Filtering (API mode uses ? prefix, SSH mode uses 'where'):")
		fmt.Println("  mt -c '/interface/print ?type=ether'")
		fmt.Println("  mt -ssh -c '/interface print where type=ether'")
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
	if !*useTLS && os.Getenv("MT_USE_TLS") == "true" {
		*useTLS = true
	}
	if !*useSSH && os.Getenv("MT_USE_SSH") == "true" {
		*useSSH = true
	}

	if *host == "" || *user == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: host, user, and password required (via flags or .env)")
		os.Exit(1)
	}

	// Set default port based on connection type
	// Only use MT_PORT from env for API mode, SSH mode defaults to 22
	if *port == "" {
		switch {
		case *useSSH:
			*port = "22"
		case *useTLS:
			*port = "8729"
		default:
			if envPort := os.Getenv("MT_PORT"); envPort != "" {
				*port = envPort
			} else {
				*port = "8728"
			}
		}
	}

	var output string
	var err error

	if *useSSH {
		output, err = runSSH(*host, *port, *user, *password, *command)
	} else {
		output, err = runAPI(*host, *port, *user, *password, *useTLS, *command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if output != "" {
		fmt.Print(output)
	}
}

func runAPI(host, port, user, password string, useTLS bool, command string) (string, error) {
	address := fmt.Sprintf("%s:%s", host, port)

	var client *routeros.Client
	var err error

	if useTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		client, err = routeros.DialTLS(address, user, password, tlsConfig)
	} else {
		client, err = routeros.Dial(address, user, password)
	}

	if err != nil {
		return "", fmt.Errorf("failed to connect to %s: %v", address, err)
	}
	defer client.Close()

	args := strings.Fields(command)
	reply, err := client.Run(args...)
	if err != nil {
		if strings.Contains(err.Error(), "!empty") {
			return "", nil
		}
		return "", err
	}

	if len(reply.Re) == 0 {
		return "", nil
	}

	var buf strings.Builder
	for _, re := range reply.Re {
		for _, pair := range re.List {
			buf.WriteString(fmt.Sprintf("%s: %s\n", pair.Key, pair.Value))
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func runSSH(host, port, user, password, command string) (string, error) {
	// Use sshpass + ssh directly to preserve $ and other special characters
	cmd := exec.Command("sshpass", "-p", password,
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "ConnectTimeout=10",
		"-p", port,
		fmt.Sprintf("%s@%s", user, host),
		command,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(output)))
		}
		return "", err
	}

	return string(output), nil
}
