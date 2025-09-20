# GophKeeper

A secure client-server password manager built with Go.

## 🚀 Quick Start

### Prerequisites
- Go 1.25.0
- PostgreSQL (optional, can use in-memory storage)

### Build and Run

```bash
# Build for current platform
make build
```

## 🔧 Configuration

### Server Configuration

The server can be configured via environment variables:

```bash
# Server settings
export SERVER_HOST=localhost
export SERVER_PORT=8080

# Database settings
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=gophkeeper
export DB_USER=postgres
export DB_PASSWORD=password
export DB_SSLMODE=disable

# JWT settings
export JWT_SECRET=your-secret-key
export JWT_TOKEN_EXPIRY=24h
```

## 📝 Usage Examples

### Server
```bash
# Start server with default settings
./build/gophkeeper-server

# Start with custom port
./build/gophkeeper-server -a localhost:9090

# Show version
./build/gophkeeper-server -version
```

### Client
```bash
# Start client
./build/gophkeeper-client

# Register new user
gophkeeper> register username password

# Login
gophkeeper> login username password

# Create data
gophkeeper> create text "My Notes" "Important notes"

# List all data
gophkeeper> list

# Get specific data
gophkeeper> get <data-id>

# Update data
gophkeeper> update <data-id>

# Delete data
gophkeeper> delete <data-id>
```