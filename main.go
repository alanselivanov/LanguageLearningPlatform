package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
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
	Role      string    `json:"role"`
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

func isValidEmail(email string) bool {
	regex := `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if user.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	if user.Password == "" || len(user.Password) < 6 {
		http.Error(w, "Password is required and must be at least 6 characters", http.StatusBadRequest)
		return
	}
	if user.Role == "" {
		user.Role = "user"
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

	if user.Name != "" && len(user.Name) < 3 {
		http.Error(w, "Name must be at least 3 characters long", http.StatusBadRequest)
		return
	}
	if user.Email != "" && !isValidEmail(user.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	if user.Password != "" && len(user.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}
	if user.Role != "" && (user.Role != "user" && user.Role != "admin") {
		http.Error(w, "Invalid role", http.StatusBadRequest)
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

func login(w http.ResponseWriter, r *http.Request) {
	var loginData struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	var user User
	if err := db.Where("name = ?", loginData.Name).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if user.Password != loginData.Password {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	response := map[string]string{
		"message": "Login successful",
		"role":    user.Role,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func sendSupportTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	file, fileHeader, err := r.FormFile("file")

	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	subject := "Support Ticket from " + name
	body := fmt.Sprintf("Name: %s\nEmail: %s\n\nMessage: %s", name, email, message)

	if err := sendEmail(subject, body, file, fileHeader); err != nil {
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ticket submitted successfully!")
}

func sendEmail(subject, body string, attachment io.Reader, fileHeader *multipart.FileHeader) error {
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := smtpUser

	if smtpUser == "" || smtpPass == "" || smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("SMTP configuration is missing. Check environment variables")
	}

	to := []string{"alan4ik.selivanov@yandex.kz"}

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\n", subject))
	msg.WriteString("MIME-Version: 1.0\n")
	msg.WriteString("Content-Type: multipart/mixed; boundary=boundary\n")
	msg.WriteString("--boundary\n")
	msg.WriteString("Content-Type: text/plain; charset=utf-8\n\n")
	msg.WriteString(body + "\n")

	if attachment != nil && fileHeader != nil {
		msg.WriteString("--boundary\n")
		msg.WriteString(fmt.Sprintf("Content-Type: application/octet-stream\n"))
		msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\n\n", fileHeader.Filename))

		fileContent, err := io.ReadAll(attachment)
		if err != nil {
			return fmt.Errorf("failed to read file content: %v", err)
		}
		msg.Write(fileContent)
		msg.WriteString("\n")
	}
	msg.WriteString("--boundary--")

	err := smtp.SendMail(smtpHost+":"+smtpPort, smtp.PlainAuth("", smtpUser, smtpPass, smtpHost), from, to, msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "main_page.html")
}

func adminPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	initDB()

	fs := http.FileServer(http.Dir("./"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/admin", adminPage)

	http.HandleFunc("/create", createUser)
	http.HandleFunc("/login", login)
	http.HandleFunc("/read", getUsers)
	http.HandleFunc("/readByID", getUserByID)
	http.HandleFunc("/update", updateUser)
	http.HandleFunc("/delete", deleteUser)

	http.HandleFunc("/send-support-ticket", sendSupportTicket)

	http.HandleFunc("/", mainPage)
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
