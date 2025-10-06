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

// CommandHandler handles CLI commands
type CommandHandler struct {
	session *client.ClientSession
	config  *client.Config
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(session *client.ClientSession, config *client.Config) *CommandHandler {
	return &CommandHandler{
		session: session,
		config:  config,
	}
}

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

	config := client.NewConfig()
	if config.ServerURL == "" {
		config.ServerURL = *serverURL
	}

	cli := client.NewClient(config.ServerURL)
	if config.Token != "" {
		cli.SetToken(config.Token)
	}

	session := client.NewClientSession(cli)
	handler := NewCommandHandler(session, config)

	runCLI(handler)
}

// runCLI runs the main CLI loop
func runCLI(handler *CommandHandler) {
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

		if handler.handleCommand(command, args) {
			break
		}
	}
}

// handleCommand processes a single command and returns true if exit was requested
func (h *CommandHandler) handleCommand(command string, args []string) bool {
	ctx := context.Background()

	switch command {
	case "register":
		return h.handleRegister(ctx, args)
	case "login":
		return h.handleLogin(ctx, args)
	case "list":
		return h.handleList(ctx)
	case "get":
		return h.handleGet(ctx, args)
	case "create":
		return h.handleCreate(ctx, args)
	case "update":
		return h.handleUpdate(ctx, args)
	case "delete":
		return h.handleDelete(ctx, args)
	case "save":
		return h.handleSave(ctx, args)
	case "help":
		h.showHelp()
		return false
	case "exit", "quit":
		fmt.Println("Goodbye!")
		return true
	default:
		fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		return false
	}
}

// handleRegister processes the register command
func (h *CommandHandler) handleRegister(ctx context.Context, args []string) bool {
	if len(args) < 2 {
		fmt.Println("Usage: register <username> <password>")
		fmt.Println("You will be prompted for a master password for data encryption")
		return false
	}
	if err := h.session.RegisterCommand(ctx, args[0], args[1], h.config); err != nil {
		fmt.Printf("Registration failed: %v\n", err)
	}
	return false
}

// handleLogin processes the login command
func (h *CommandHandler) handleLogin(ctx context.Context, args []string) bool {
	if len(args) < 2 {
		fmt.Println("Usage: login <username> <password>")
		fmt.Println("You will be prompted for your master password")
		return false
	}
	if err := h.session.LoginCommand(ctx, args[0], args[1], h.config); err != nil {
		fmt.Printf("Login failed: %v\n", err)
	}
	return false
}

// handleList processes the list command
func (h *CommandHandler) handleList(ctx context.Context) bool {
	if err := h.session.ListCommand(ctx); err != nil {
		if err == client.ErrNotAuthenticated {
			fmt.Println("Please login first to access encrypted data")
		} else {
			fmt.Printf("Failed to list data: %v\n", err)
		}
	}
	return false
}

// handleGet processes the get command
func (h *CommandHandler) handleGet(ctx context.Context, args []string) bool {
	if len(args) < 1 {
		fmt.Println("Usage: get <id>")
		return false
	}
	if err := h.session.GetCommand(ctx, args[0]); err != nil {
		if err == client.ErrNotAuthenticated {
			fmt.Println("Please login first to access encrypted data")
		} else {
			fmt.Printf("Failed to get data: %v\n", err)
		}
	}
	return false
}

// handleCreate processes the create command
func (h *CommandHandler) handleCreate(ctx context.Context, args []string) bool {
	if len(args) < 2 {
		fmt.Println("Usage: create <type> <name> [description]")
		fmt.Println("Types: login_password, text, binary, bank_card")
		fmt.Println("Note: Use quotes around names with spaces: create text \"My Shopping List\" \"Description\"")
		return false
	}
	description := ""
	if len(args) > 2 {
		description = client.CleanQuotes(strings.Join(args[2:], " "))
	}
	if err := h.session.CreateCommand(ctx, args[0], client.CleanQuotes(args[1]), description); err != nil {
		if err == client.ErrNotAuthenticated {
			fmt.Println("Please login first to create encrypted data")
		} else {
			fmt.Printf("Failed to create data: %v\n", err)
		}
	}
	return false
}

// handleUpdate processes the update command
func (h *CommandHandler) handleUpdate(ctx context.Context, args []string) bool {
	if len(args) < 1 {
		fmt.Println("Usage: update <id>")
		return false
	}
	if err := h.session.UpdateCommand(ctx, args[0]); err != nil {
		if err == client.ErrNotAuthenticated {
			fmt.Println("Please login first to update encrypted data")
		} else {
			fmt.Printf("Failed to update data: %v\n", err)
		}
	}
	return false
}

// handleDelete processes the delete command
func (h *CommandHandler) handleDelete(ctx context.Context, args []string) bool {
	if len(args) < 1 {
		fmt.Println("Usage: delete <id>")
		return false
	}
	if err := h.session.DeleteCommand(ctx, args[0]); err != nil {
		if err == client.ErrNotAuthenticated {
			fmt.Println("Please login first to delete encrypted data")
		} else {
			fmt.Printf("Failed to delete data: %v\n", err)
		}
	}
	return false
}

// handleSave processes the save command
func (h *CommandHandler) handleSave(ctx context.Context, args []string) bool {
	if len(args) < 1 {
		fmt.Println("Usage: save <id> [output_path]")
		fmt.Println("Note: This command only works with binary data types")
		return false
	}
	outputPath := ""
	if len(args) > 1 {
		outputPath = args[1]
	}
	if err := h.session.SaveCommand(ctx, args[0], outputPath); err != nil {
		if err == client.ErrNotAuthenticated {
			fmt.Println("Please login first to save encrypted data")
		} else {
			fmt.Printf("Failed to save data: %v\n", err)
		}
	}
	return false
}

// showHelp displays help information from file
func (h *CommandHandler) showHelp() {
	content, err := os.ReadFile("assets/client/help.txt")
	if err != nil {
		fmt.Println("Error reading help file:", err)
		return
	}

	fmt.Print(string(content))
}
