package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a2sh3r/gophkeeper/internal/client"
	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/a2sh3r/gophkeeper/internal/models"
	"github.com/a2sh3r/gophkeeper/pkg/version"
	"go.uber.org/zap"
)

const (
	configFile = ".gophkeeper_config"
)

type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
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

	config := loadConfig()
	if config.ServerURL == "" {
		config.ServerURL = *serverURL
	}

	cli := client.NewClient(config.ServerURL)
	if config.Token != "" {
		cli.SetToken(config.Token)
	}

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

		switch command {
		case "register":
			handleRegister(cli, args, config)
		case "login":
			handleLogin(cli, args, config)
		case "list":
			handleList(cli)
		case "get":
			handleGet(cli, args)
		case "create":
			handleCreate(cli, args)
		case "update":
			handleUpdate(cli, args)
		case "delete":
			handleDelete(cli, args)
		case "save":
			handleSave(cli, args)
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

func loadConfig() *Config {
	config := &Config{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Log.Error("Failed to return home dir", zap.Error(err))
		return config
	}

	configPath := fmt.Sprintf("%s/%s", homeDir, configFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Log.Error("Failed to read file", zap.Error(err))
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		logger.Log.Error("Failed to unmarshal data", zap.Error(err))
		return config
	}
	return config
}

func saveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Log.Error("Failed to return home dir", zap.Error(err))
		return err
	}

	configPath := fmt.Sprintf("%s/%s", homeDir, configFile)
	data, err := json.Marshal(config)
	if err != nil {
		logger.Log.Error("Failed to marshal data", zap.Error(err))
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func handleRegister(cli *client.Client, args []string, config *Config) {
	if len(args) < 2 {
		fmt.Println("Usage: register <username> <password>")
		return
	}

	username := args[0]
	password := args[1]

	resp, err := cli.Register(username, password)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}

	config.Token = resp.Token
	if err := saveConfig(config); err != nil {
		logger.Log.Error("Failed to return config", zap.Error(err))
	}
	cli.SetToken(resp.Token)

	fmt.Printf("Successfully registered user: %s\n", resp.User.Username)
}

func handleLogin(cli *client.Client, args []string, config *Config) {
	if len(args) < 2 {
		fmt.Println("Usage: login <username> <password>")
		return
	}

	username := args[0]
	password := args[1]

	resp, err := cli.Login(username, password)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	config.Token = resp.Token
	if err := saveConfig(config); err != nil {
		logger.Log.Error("Failed to save config", zap.Error(err))
	}
	cli.SetToken(resp.Token)

	fmt.Printf("Successfully logged in as: %s\n", resp.User.Username)
}

func handleList(cli *client.Client) {
	data, err := cli.GetData()
	if err != nil {
		fmt.Printf("Failed to get data: %v\n", err)
		return
	}

	if len(data) == 0 {
		fmt.Println("No data found")
		return
	}

	fmt.Printf("Found %d items:\n", len(data))
	for _, item := range data {
		fmt.Printf("  %s [%s] - %s", item.ID.String(), item.Type, cleanQuotes(item.Name))
		if item.Description != "" {
			fmt.Printf(" - %s", cleanQuotes(item.Description))
		}
		fmt.Printf("\n")
	}
}

func handleGet(cli *client.Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: get <id>")
		return
	}

	id := args[0]
	data, err := cli.GetDataByID(id)
	if err != nil {
		fmt.Printf("Failed to get data: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", data.ID)
	fmt.Printf("Type: %s\n", data.Type)
	fmt.Printf("Name: %s\n", cleanQuotes(data.Name))
	fmt.Printf("Description: %s\n", cleanQuotes(data.Description))
	fmt.Printf("Created: %s\n", data.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", data.UpdatedAt.Format("2006-01-02 15:04:05"))

	if data.Metadata != "" {
		fmt.Printf("Metadata: %s\n", data.Metadata)
	}

	displayStructuredData(data)
}

// displayStructuredData displays data in a structured format based on type
func displayStructuredData(data *models.Data) {
	switch data.Type {
	case "login_password":
		var loginData models.LoginPasswordData
		if err := json.Unmarshal(data.Data, &loginData); err == nil {
			fmt.Printf("Login: %s\n", loginData.Login)
			fmt.Printf("Password: %s\n", loginData.Password)
			if loginData.URL != "" {
				fmt.Printf("URL: %s\n", loginData.URL)
			}
			if loginData.Notes != "" {
				fmt.Printf("Notes: %s\n", loginData.Notes)
			}
		} else {
			fmt.Printf("Data: %s\n", string(data.Data))
		}
	case "text":
		var textData models.TextData
		if err := json.Unmarshal(data.Data, &textData); err == nil {
			fmt.Printf("Content: %s\n", textData.Content)
			if textData.Notes != "" {
				fmt.Printf("Notes: %s\n", textData.Notes)
			}
		} else {
			fmt.Printf("Data: %s\n", string(data.Data))
		}
	case "binary":
		var binaryData models.BinaryData
		if err := json.Unmarshal([]byte(data.Metadata), &binaryData); err == nil {
			fmt.Printf("File Name: %s\n", binaryData.FileName)
			fmt.Printf("MIME Type: %s\n", binaryData.MimeType)
			fmt.Printf("Size: %d bytes\n", binaryData.Size)
			if binaryData.Notes != "" {
				fmt.Printf("Notes: %s\n", binaryData.Notes)
			}
			fmt.Printf("File Data: [Base64 encoded binary data - %d bytes]\n", len(data.Data))
		} else {
			fmt.Printf("Data: %s\n", string(data.Data))
		}
	case "bank_card":
		var bankCardData models.BankCardData
		if err := json.Unmarshal(data.Data, &bankCardData); err == nil {
			fmt.Printf("Card Number: %s\n", bankCardData.CardNumber)
			fmt.Printf("Expiry Date: %s\n", bankCardData.ExpiryDate)
			fmt.Printf("CVV: %s\n", bankCardData.CVV)
			fmt.Printf("Cardholder: %s\n", bankCardData.Cardholder)
			if bankCardData.Bank != "" {
				fmt.Printf("Bank: %s\n", bankCardData.Bank)
			}
			if bankCardData.Notes != "" {
				fmt.Printf("Notes: %s\n", bankCardData.Notes)
			}
		} else {
			fmt.Printf("Data: %s\n", string(data.Data))
		}
	default:
		fmt.Printf("Data: %s\n", string(data.Data))
	}
}

func handleCreate(cli *client.Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: create <type> <name> [description]")
		fmt.Println("Types: login_password, text, binary, bank_card")
		fmt.Println("Note: Use quotes around names with spaces: create text \"My Shopping List\" \"Description\"")
		return
	}

	dataType := args[0]
	name := cleanQuotes(args[1])
	description := ""
	if len(args) > 2 {
		description = cleanQuotes(strings.Join(args[2:], " "))
	}

	var dataContent []byte
	var metadata string

	switch dataType {
	case "login_password":
		dataContent, metadata = createLoginPasswordData()
	case "text":
		dataContent, metadata = createTextData()
	case "binary":
		dataContent, metadata = createBinaryData()
	case "bank_card":
		dataContent, metadata = createBankCardData()
	default:
		fmt.Printf("Unknown data type: %s\n", dataType)
		return
	}

	dataReq := models.DataRequest{
		Type:        models.DataType(dataType),
		Name:        name,
		Description: description,
		Data:        dataContent,
		Metadata:    metadata,
	}

	data, err := cli.CreateData(dataReq)
	if err != nil {
		fmt.Printf("Failed to create data: %v\n", err)
		return
	}

	fmt.Printf("Successfully created data with ID: %s\n", data.ID)
}

func handleUpdate(cli *client.Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: update <id>")
		return
	}

	id := args[0]

	data, err := cli.GetDataByID(id)
	if err != nil {
		fmt.Printf("Failed to get data: %v\n", err)
		return
	}

	fmt.Printf("Current data: %s\n", string(data.Data))
	fmt.Print("Enter new data content: ")
	scanner := bufio.NewScanner(os.Stdin)
	var newContent string
	if scanner.Scan() {
		newContent = scanner.Text()
	}

	dataReq := models.DataRequest{
		Type:        data.Type,
		Name:        data.Name,
		Description: data.Description,
		Data:        []byte(newContent),
		Metadata:    data.Metadata,
	}

	updatedData, err := cli.UpdateData(id, dataReq)
	if err != nil {
		fmt.Printf("Failed to update data: %v\n", err)
		return
	}

	fmt.Printf("Successfully updated data: %s\n", updatedData.ID)
}

func handleDelete(cli *client.Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: delete <id>")
		return
	}

	id := args[0]

	fmt.Printf("Are you sure you want to delete data %s? (y/N): ", id)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	confirmation := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if confirmation != "y" && confirmation != "yes" {
		fmt.Println("Deletion cancelled")
		return
	}

	err := cli.DeleteData(id)
	if err != nil {
		fmt.Printf("Failed to delete data: %v\n", err)
		return
	}

	fmt.Printf("Successfully deleted data: %s\n", id)
}

// handleSave saves binary data to a file
func handleSave(cli *client.Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: save <id> [output_path]")
		fmt.Println("Note: This command only works with binary data types")
		return
	}

	id := args[0]

	// Get data from server
	data, err := cli.GetDataByID(id)
	if err != nil {
		fmt.Printf("Failed to get data: %v\n", err)
		return
	}

	// Check if it's binary data
	if data.Type != "binary" {
		fmt.Printf("Error: Data with ID %s is not binary type (type: %s)\n", id, data.Type)
		fmt.Println("The save command only works with binary data types")
		return
	}

	// Parse binary metadata
	var binaryData models.BinaryData
	if err := json.Unmarshal([]byte(data.Metadata), &binaryData); err != nil {
		fmt.Printf("Failed to parse binary metadata: %v\n", err)
		return
	}

	// Determine output path
	var outputPath string
	if len(args) > 1 {
		outputPath = args[1]
	} else {
		// Use original filename if no path specified
		outputPath = binaryData.FileName
	}

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("File %s already exists. Overwrite? (y/N): ", outputPath)
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return
		}
		confirmation := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if confirmation != "y" && confirmation != "yes" {
			fmt.Println("Save cancelled")
			return
		}
	}

	// Decode base64 data
	fileData, err := base64.StdEncoding.DecodeString(string(data.Data))
	if err != nil {
		fmt.Printf("Failed to decode binary data: %v\n", err)
		return
	}

	// Write file
	err = os.WriteFile(outputPath, fileData, 0644)
	if err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
		return
	}

	fmt.Printf("Successfully saved binary data to: %s\n", outputPath)
	fmt.Printf("File: %s\n", binaryData.FileName)
	fmt.Printf("Size: %d bytes\n", binaryData.Size)
	fmt.Printf("MIME Type: %s\n", binaryData.MimeType)
	if binaryData.Notes != "" {
		fmt.Printf("Notes: %s\n", binaryData.Notes)
	}
}

func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  register <username> <password>  - Register a new user")
	fmt.Println("  login <username> <password>     - Login with existing user")
	fmt.Println("  list                            - List all data")
	fmt.Println("  get <id>                        - Get data by ID")
	fmt.Println("  create <type> <name> [desc]     - Create new data")
	fmt.Println("  update <id>                     - Update existing data")
	fmt.Println("  delete <id>                     - Delete data")
	fmt.Println("  save <id> [path]                - Save binary data to file")
	fmt.Println("  help                            - Show this help")
	fmt.Println("  exit, quit                      - Exit the program")
	fmt.Println()
	fmt.Println("Data types:")
	fmt.Println("  login_password - Login/password pairs with URL and notes")
	fmt.Println("  text          - Arbitrary text data with notes")
	fmt.Println("  binary        - Binary files (PDF, images, documents, etc.)")
	fmt.Println("  bank_card     - Bank card data (number, expiry, CVV, holder)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  create text \"Shopping List\" \"Weekly groceries\"")
	fmt.Println("  create login_password \"GitHub\" \"Development account\"")
	fmt.Println("  create bank_card \"Visa Card\" \"Main credit card\"")
	fmt.Println("  create binary \"Contract\" \"Important document\"")
	fmt.Println("  save <id> /path/to/save/file.pdf")
	fmt.Println()
	fmt.Println("Note: Use quotes around names and descriptions that contain spaces!")
	fmt.Println("Binary files: Enter full file path when prompted")
	fmt.Println("Save command: Downloads binary data to local file system")
	fmt.Println("All data types support metadata and structured storage!")
}

func cleanQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// createLoginPasswordData creates structured login/password data
func createLoginPasswordData() ([]byte, string) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter login/username: ")
	scanner.Scan()
	login := scanner.Text()

	fmt.Print("Enter password: ")
	scanner.Scan()
	password := scanner.Text()

	fmt.Print("Enter URL/website (optional): ")
	scanner.Scan()
	url := scanner.Text()

	fmt.Print("Enter notes (optional): ")
	scanner.Scan()
	notes := scanner.Text()

	loginData := models.LoginPasswordData{
		Login:    login,
		Password: password,
		URL:      url,
		Notes:    notes,
	}

	jsonData, _ := json.Marshal(loginData)
	metadata := fmt.Sprintf("website:%s", url)

	return jsonData, metadata
}

// createTextData creates structured text data
func createTextData() ([]byte, string) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter text content: ")
	scanner.Scan()
	content := scanner.Text()

	fmt.Print("Enter notes (optional): ")
	scanner.Scan()
	notes := scanner.Text()

	textData := models.TextData{
		Content: content,
		Notes:   notes,
	}

	jsonData, _ := json.Marshal(textData)
	metadata := fmt.Sprintf("notes:%s", notes)

	return jsonData, metadata
}

// createBinaryData creates structured binary data
func createBinaryData() ([]byte, string) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter file path: ")
	scanner.Scan()
	filePath := scanner.Text()

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return []byte{}, ""
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return []byte{}, ""
	}

	fmt.Print("Enter notes (optional): ")
	scanner.Scan()
	notes := scanner.Text()

	mimeType := getMimeType(filepath.Ext(filePath))

	binaryData := models.BinaryData{
		FileName: filepath.Base(filePath),
		MimeType: mimeType,
		Size:     fileInfo.Size(),
		Notes:    notes,
	}

	jsonData, _ := json.Marshal(binaryData)
	metadata := string(jsonData)

	return []byte(base64.StdEncoding.EncodeToString(fileData)), metadata
}

func createBankCardData() ([]byte, string) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter card number: ")
	scanner.Scan()
	cardNumber := scanner.Text()

	fmt.Print("Enter expiry date (MM/YY): ")
	scanner.Scan()
	expiryDate := scanner.Text()

	fmt.Print("Enter CVV: ")
	scanner.Scan()
	cvv := scanner.Text()

	fmt.Print("Enter cardholder name: ")
	scanner.Scan()
	cardholder := scanner.Text()

	fmt.Print("Enter bank name (optional): ")
	scanner.Scan()
	bank := scanner.Text()

	fmt.Print("Enter notes (optional): ")
	scanner.Scan()
	notes := scanner.Text()

	bankCardData := models.BankCardData{
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		CVV:        cvv,
		Cardholder: cardholder,
		Bank:       bank,
		Notes:      notes,
	}

	jsonData, _ := json.Marshal(bankCardData)
	metadata := fmt.Sprintf("bank:%s,cardholder:%s", bank, cardholder)

	return jsonData, metadata
}

func getMimeType(ext string) string {
	switch strings.ToLower(ext) {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	default:
		return "application/octet-stream"
	}
}
