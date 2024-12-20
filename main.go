package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var db *sql.DB

func initDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var errDB error
	db, errDB = sql.Open("postgres", psqlInfo)
	if errDB != nil {
		log.Fatal(errDB)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	fmt.Println("Successfully connected to the database!")
}
func getUserByID(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var user User
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		fmt.Println("Error decoding JSON:", err)
		return
	}

	fmt.Printf("Received User: %+v\n", user)

	query := `INSERT INTO users (name, email, password, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = db.QueryRow(query, user.Name, user.Email, user.Password, time.Now(), time.Now()).Scan(&user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error executing query:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, name, email, password, created_at, updated_at FROM users`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
		users = append(users, user)
	}

	json.NewEncoder(w).Encode(users)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		fmt.Println("Error decoding JSON:", err)
		return
	}

	fmt.Printf("Received User for Update: %+v\n", user)

	if user.ID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		fmt.Println("Invalid user ID:", user.ID)
		return
	}

	query := `UPDATE users SET name = $1, email = $2, password = $3, updated_at = $4 WHERE id = $5`
	result, err := db.Exec(query, user.Name, user.Email, user.Password, time.Now(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error executing query:", err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error getting rows affected:", err)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		fmt.Println("No rows affected, user ID not found:", user.ID)
		return
	}

	fmt.Printf("User with ID %d updated successfully\n", user.ID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		fmt.Println("Error decoding JSON:", err)
		return
	}

	fmt.Printf("Received ID for deletion: %d\n", user.ID)

	query := `DELETE FROM users WHERE id = $1`
	result, err := db.Exec(query, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error executing query:", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		fmt.Println("User not found for ID:", user.ID)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted"})
}

func logRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

func main() {
	initDB()
	defer db.Close()

	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", logRequests(fs))

	http.HandleFunc("/create", createUser)
	http.HandleFunc("/read", getUsers)
	http.HandleFunc("/update", updateUser)
	http.HandleFunc("/delete", deleteUser)
	http.HandleFunc("/readByID", getUserByID)

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
