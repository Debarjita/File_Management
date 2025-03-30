# File Sharing Platform (Go Backend)

This is a file-sharing platform backend built using Go, Gin framework, and JWT authentication.

## Features
- User authentication (Register/Login)
- JWT-based authentication middleware
- File upload and download
- User management
- Secure password hashing with bcrypt

## Tech Stack
- **Go** (Golang)
- **Gin** (Web framework)
- **JWT** (Authentication)
- **bcrypt** (Password hashing)
- **PostgreSQL** (Database)

## Installation
1. Clone the repository:
   ```sh
   git clone https://github.com/your-username/your-repo.git
   ```
2. Navigate to the project folder:
   ```sh
   cd file-sharing-platform
   ```
3. Install dependencies:
   ```sh
   go mod tidy
   ```
4. Set up environment variables:
   - Create a `.env` file and configure database and JWT secret.
   ```env
   DATABASE_URL=your_database_url
   JWT_SECRET=your_secret_key
   ```

## Running the Project
```sh
go run cmd/main.go
```

## API Endpoints

### Authentication
| Method | Endpoint          | Description          |
|--------|------------------|----------------------|
| POST   | `/api/register`  | Register new user   |
| POST   | `/api/login`     | Login user          |

### File Management
| Method | Endpoint          | Description          |
|--------|------------------|----------------------|
| POST   | `/api/upload`    | Upload a file       |
| GET    | `/api/download`  | Download a file     |

## Folder Structure
```
file-sharing-platform/
├── cmd/
│   ├── main.go
├── internal/
│   ├── api/
│   │   ├── auth_handler.go
│   ├── db/
│   │   ├── database.go
│   ├── models/
│   │   ├── model.go
│   ├── auth/
│   │   ├── jwt.go
│   ├── middleware/
│   │   ├── middleware.go
├── go.mod
├── go.sum
└── README.md
```

## Contributing
Feel free to fork and contribute by submitting a pull request.

## License
This project is licensed under the MIT License.
