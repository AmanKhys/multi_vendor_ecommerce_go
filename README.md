# Multi-Vendor Ecommerce Application

This is a multi-vendor eCommerce application built in Go, designed with a modular monolithic architecture. It leverages the standard `net/http` package for handling HTTP requests and uses PostgreSQL as the database server.

## Features
- Modular monolithic architecture
- Built using Go with `net/http` (version 1.22) and a framework
- Uses PostgreSQL for database management
- SQLC for query generation and schema management
- Custom database functions instead of GORM

## Installation & Setup

### Prerequisites
Ensure you have the following installed:
- Go (Version: 1.22+)
- PostgreSQL
- SQLC

### Running the Application

Clone the repository and navigate to the project directory:
```sh
git clone github.com/amankhys/multi_vendor_ecommerce_go.git
cd multi_vendor_ecommerce_go
```

Run the application using:
```sh
make  # If using Makefile
```
OR
```sh
go run cmd/main.go  # Direct execution
```

## Database Setup
Ensure that PostgreSQL is running and configured properly. Use SQLC to generate database queries and schema before running the application.

## Contributing
Contributions are welcome! Feel free to open issues or submit pull requests.

