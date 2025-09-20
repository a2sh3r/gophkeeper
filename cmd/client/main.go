package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/a2sh3r/gophkeeper/internal/client"
	"github.com/a2sh3r/gophkeeper/pkg/version"
)

func main() {
	var (
		serverURL   = flag.String("server", "http://localhost:8080", "Server URL")
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Info())
		os.Exit(0)
	}

	config := client.LoadConfig()
	if config.ServerURL == "" {
		config.ServerURL = *serverURL
	}

	cli := client.NewClient(config.ServerURL)
	if config.Token != "" {
		cli.SetToken(config.Token)
	}

	session := client.NewClientSession(cli)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("gophkeeper> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		ctx := context.Background()

		switch command {
		case "register":
			if len(args) < 2 {
				fmt.Println("Usage: register <username> <password>")
				fmt.Println("You will be prompted for a master password for data encryption")
				continue
			}
			if err := session.RegisterCommand(ctx, args[0], args[1], config); err != nil {
				fmt.Printf("Registration failed: %v\n", err)
			}

		case "login":
			if len(args) < 2 {
				fmt.Println("Usage: login <username> <password>")
				fmt.Println("You will be prompted for your master password")
				continue
			}
			if err := session.LoginCommand(ctx, args[0], args[1], config); err != nil {
				fmt.Printf("Login failed: %v\n", err)
			}

		case "list":
			if err := session.ListCommand(ctx); err != nil {
				if err == client.ErrNotAuthenticated {
					fmt.Println("Please login first to access encrypted data")
				} else {
					fmt.Printf("Failed to list data: %v\n", err)
				}
			}

		case "get":
			if len(args) < 1 {
				fmt.Println("Usage: get <id>")
				continue
			}
			if err := session.GetCommand(ctx, args[0]); err != nil {
				if err == client.ErrNotAuthenticated {
					fmt.Println("Please login first to access encrypted data")
				} else {
					fmt.Printf("Failed to get data: %v\n", err)
				}
			}

		case "create":
			if len(args) < 2 {
				fmt.Println("Usage: create <type> <name> [description]")
				fmt.Println("Types: login_password, text, binary, bank_card")
				fmt.Println("Note: Use quotes around names with spaces: create text \"My Shopping List\" \"Description\"")
				continue
			}
			description := ""
			if len(args) > 2 {
				description = client.CleanQuotes(strings.Join(args[2:], " "))
			}
			if err := session.CreateCommand(ctx, args[0], client.CleanQuotes(args[1]), description); err != nil {
				if err == client.ErrNotAuthenticated {
					fmt.Println("Please login first to create encrypted data")
				} else {
					fmt.Printf("Failed to create data: %v\n", err)
				}
			}

		case "update":
			if len(args) < 1 {
				fmt.Println("Usage: update <id>")
				continue
			}
			if err := session.UpdateCommand(ctx, args[0]); err != nil {
				if err == client.ErrNotAuthenticated {
					fmt.Println("Please login first to update encrypted data")
				} else {
					fmt.Printf("Failed to update data: %v\n", err)
				}
			}

		case "delete":
			if len(args) < 1 {
				fmt.Println("Usage: delete <id>")
				continue
			}
			if err := session.DeleteCommand(ctx, args[0]); err != nil {
				if err == client.ErrNotAuthenticated {
					fmt.Println("Please login first to delete encrypted data")
				} else {
					fmt.Printf("Failed to delete data: %v\n", err)
				}
			}

		case "save":
			if len(args) < 1 {
				fmt.Println("Usage: save <id> [output_path]")
				fmt.Println("Note: This command only works with binary data types")
				continue
			}
			outputPath := ""
			if len(args) > 1 {
				outputPath = args[1]
			}
			if err := session.SaveCommand(ctx, args[0], outputPath); err != nil {
				if err == client.ErrNotAuthenticated {
					fmt.Println("Please login first to save encrypted data")
				} else {
					fmt.Printf("Failed to save data: %v\n", err)
				}
			}

		case "help":
			showHelp()

		case "exit", "quit":
			fmt.Println("Goodbye!")
			os.Exit(0)

		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		}
	}
}

func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  register <username> <password>  - Register a new user (requires master password)")
	fmt.Println("  login <username> <password>     - Login with existing user (requires master password)")
	fmt.Println("  list                            - List all encrypted data")
	fmt.Println("  get <id>                        - Get and decrypt data by ID")
	fmt.Println("  create <type> <name> [desc]     - Create new encrypted data")
	fmt.Println("  update <id>                     - Update existing encrypted data")
	fmt.Println("  delete <id>                     - Delete encrypted data")
	fmt.Println("  save <id> [path]                - Save decrypted binary data to file")
	fmt.Println("  help                            - Show this help")
	fmt.Println("  exit, quit                      - Exit the program")
	fmt.Println()
	fmt.Println("Data types (all encrypted):")
	fmt.Println("  login_password - Login/password pairs with URL and notes")
	fmt.Println("  text          - Arbitrary text data with notes")
	fmt.Println("  binary        - Binary files (PDF, images, documents, etc.)")
	fmt.Println("  bank_card     - Bank card data (number, expiry, CVV, holder)")
	fmt.Println()
	fmt.Println("Security features:")
	fmt.Println("  🔐 End-to-end encryption with AES-256-GCM")
	fmt.Println("  🔑 Master password required for all data operations")
	fmt.Println("  🛡️  Data encrypted on client before sending to server")
	fmt.Println("  🔒 Server never sees unencrypted data")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  create text \"Shopping List\" \"My grocery list\"")
	fmt.Println("  create login_password \"Gmail Account\" \"My Gmail login\"")
	fmt.Println("  create binary \"Important Document.pdf\" \"Contract document\"")
	fmt.Println("  create bank_card \"Visa Card\" \"My primary credit card\"")
	fmt.Println("  get 123e4567-e89b-12d3-a456-426614174000")
	fmt.Println("  save 123e4567-e89b-12d3-a456-426614174000 ./downloaded_file.pdf")
}
