package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os" // To read environment variables

	_ "github.com/lib/pq" // PostgreSQL driver
)

// User struct to hold user data (example)
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db *sql.DB // Global database connection pool

func main() {
	// 1. Get database connection details from environment variables
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")

	// Construct the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Attempting to connect to database: %s", connStr)

	var err error
	// 2. Establish database connection
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close() // Ensure the connection is closed when main exits

	// 3. Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database (ping failed): %v", err)
	}
	log.Println("Successfully connected to the PostgreSQL database!")

	// 4. Create a table if it doesn't exist (simple example)
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
	}
	log.Println("Users table checked/created successfully.")

	// 5. Set up HTTP routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/users", usersHandler) // Handle both GET and POST for users

	// Start the server
	port := os.Getenv("APP_PORT") // You might define APP_PORT in docker-compose.yml or default
	if port == "" {
		port = "8080" // Default port
	}
	listenAddr := fmt.Sprintf(":%s", port)
	fmt.Printf("Go application listening on %s\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

// homeHandler handles the root path
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from Go in Docker!")
}

// healthHandler checks database connection health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	err := db.Ping()
	if err != nil {
		http.Error(w, fmt.Sprintf("Database connection failed: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Service is healthy and connected to DB!")
}

// usersHandler handles GET and POST requests for users
func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getUsers(w, r)
	case "POST":
		createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getUsers retrieves all users from the database
func getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying users: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// createUser inserts a new user into the database
func createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation (e.g., name and email not empty)
	if u.Name == "" || u.Email == "" {
		http.Error(w, "Name and Email are required", http.StatusBadRequest)
		return
	}

	// Insert the user into the database
	sqlStatement := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`
	var newID int
	err = db.QueryRow(sqlStatement, u.Name, u.Email).Scan(&newID)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			http.Error(w, "User with this email already exists", http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Error inserting new user: %v", err), http.StatusInternalServerError)
		}
		return
	}

	u.ID = newID // Assign the new ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}
