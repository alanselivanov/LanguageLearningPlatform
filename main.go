package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var db *gorm.DB

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

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("Database connected and migrated successfully!")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := db.Create(&user).Error; err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		http.Error(w, "Error retrieving users", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func getUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var user User
	if err := db.First(&user, id).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	if user.ID == 0 {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	user.UpdatedAt = time.Now()

	if err := db.Model(&User{}).Where("id = ?", user.ID).Updates(user).Error; err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	if user.ID == 0 {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if err := db.Delete(&User{}, user.ID).Error; err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	initDB()
	http.HandleFunc("/create", createUser)
	http.HandleFunc("/read", getUsers)
	http.HandleFunc("/readByID", getUserByID)
	http.HandleFunc("/update", updateUser)
	http.HandleFunc("/delete", deleteUser)

	http.HandleFunc("/", serveIndex)
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
