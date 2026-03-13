# Tasks APIs 
## Description
A simple CRUD with JWT Auth using Golang, Gin and Gorm.

## System Design
Please check [here](./docs/system-design.md) to see more detail about the APIs system design

## Project Setup
### 1. Prerequisites
* **Go** (version 1.21 or higher)
* **PostgreSQL**
* **Make** (optional, for using the Makefile shortcuts)

#### 2. Set Up Environment Variables

Create a `.env` file in the root directory with the following configuration:

```env
DB_USER=root
DB_PASSWORD=password
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=tasks_api
JWT_SECRET=your_secret_key
PORT=8000
```
Check the `.env-example` file [here](.env-example).

### 3. Install Dependencies

```bash
go mod tidy
````

---

## Running the Project

You can run the project directly using Go or via the provided Makefile.

### Development Mode

```bash
# Using Makefile
make run

# Or using Go directly
go run main.go
```

The application will automatically run database migrations and seeds upon startup.

---

## Running Tests

This project follows Go's idiomatic testing patterns, including table-driven tests and environment isolation.

### Run All Tests

```bash
make test
```

### Run Test Coverage

To generate a coverage report and view it in your browser:

```bash
make test-coverage
```

### Clean Up

To remove generated coverage files:

```bash
make clean
```
---
## Documentation
### Postman
**Note:** <br>
Check the postman collection [here](Golang_Books_CRUD.postman_collection.json)

**Steps to Import Postman Collection**:
1. Open Postman.
2. Click on "Import" in the top-left corner.
3. Select "File" and upload the `Golang_Books_CRUD.postman_collection.json` file from this repository.
4. After importing, you can start making API requests using the defined endpoints.
---
## Additional Notes
- **Environment Variables**: The `.env` file is critical for configuring the application correctly, especially for connecting to the PostgreSQL database and generating JWT tokens.