package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
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

var (
	db      *gorm.DB
	logger  *logrus.Logger
	limiter = rate.NewLimiter(1, 3)
)

func initLogger() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
}

func logUserAction(action, status string, details map[string]interface{}) {
	logEntry := logger.WithFields(logrus.Fields{
		"action":  action,
		"status":  status,
		"details": details,
		"time":    time.Now(),
	})

	if status == "success" {
		logEntry.Info("User Action")
	} else if status == "error" {
		logEntry.Error("User Action")
	} else {
		logEntry.Warn("User Action")
	}
}

func logClientError(w http.ResponseWriter, r *http.Request) {
	var errorDetails map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&errorDetails); err != nil {
		handleError(w, "logClientError", fmt.Errorf("invalid error details: %v", err), http.StatusBadRequest)
		return
	}

	logEntry := logger.WithFields(logrus.Fields{
		"action":  "logClientError",
		"status":  "error",
		"details": errorDetails,
		"time":    time.Now(),
	})

	logEntry.Error("Client-side error logged")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Error logged successfully"})
}

func rateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			logUserAction("rateLimiter", "error", map[string]interface{}{
				"ip":     r.RemoteAddr,
				"reason": "Rate limit exceeded",
			})
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func initDB() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to the database:", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		logger.Fatal("Failed to migrate database:", err)
	}

	logger.Info("Database connected and migrated successfully!")
}

func isValidEmail(email string) bool {
	regex := `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}

func handleError(w http.ResponseWriter, action string, err error, statusCode int) {
	http.Error(w, err.Error(), statusCode)
	logUserAction(action, "error", map[string]interface{}{"error": err.Error()})
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		handleError(w, "createUser", fmt.Errorf("invalid input data: %v", err), http.StatusBadRequest)
		return
	}

	// Логирование входных данных для отладки
	logger.WithFields(logrus.Fields{
		"name":     user.Name,
		"email":    user.Email,
		"password": user.Password,
	}).Info("Received createUser request")

	if user.Name == "" {
		handleError(w, "createUser", fmt.Errorf("name is required"), http.StatusBadRequest)
		return
	}
	if user.Email == "" {
		handleError(w, "createUser", fmt.Errorf("email is required"), http.StatusBadRequest)
		return
	}
	if !isValidEmail(user.Email) {
		handleError(w, "createUser", fmt.Errorf("invalid email format"), http.StatusBadRequest)
		return
	}
	if user.Password == "" || len(user.Password) < 6 {
		handleError(w, "createUser", fmt.Errorf("password must be at least 6 characters"), http.StatusBadRequest)
		return
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := db.Create(&user).Error; err != nil {
		handleError(w, "createUser", fmt.Errorf("error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
	logUserAction("createUser", "success", map[string]interface{}{"user_id": user.ID})
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		handleError(w, "getUsers", fmt.Errorf("error retrieving users: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
	logUserAction("getUsers", "success", map[string]interface{}{"count": len(users)})
}

func getUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		handleError(w, "getUserByID", fmt.Errorf("id is required"), http.StatusBadRequest)
		return
	}

	var user User
	if err := db.First(&user, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			handleError(w, "getUserByID", fmt.Errorf("user not found: %v", err), http.StatusNotFound)
			return
		}

		handleError(w, "getUserByID", fmt.Errorf("database error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
	logUserAction("getUserByID", "success", map[string]interface{}{
		"user": user,
	})
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		handleError(w, "updateUser", fmt.Errorf("invalid input data: %v", err), http.StatusBadRequest)
		return
	}

	if user.ID == 0 {
		handleError(w, "updateUser", fmt.Errorf("user ID is required"), http.StatusBadRequest)
		return
	}

	if user.Name != "" && len(user.Name) < 3 {
		handleError(w, "updateUser", fmt.Errorf("name must be at least 3 characters long"), http.StatusBadRequest)
		return
	}

	if user.Email != "" && !isValidEmail(user.Email) {
		handleError(w, "updateUser", fmt.Errorf("invalid email format"), http.StatusBadRequest)
		return
	}

	if user.Password != "" && len(user.Password) < 6 {
		handleError(w, "updateUser", fmt.Errorf("password must be at least 6 characters"), http.StatusBadRequest)
		return
	}

	user.UpdatedAt = time.Now()

	if err := db.Model(&User{}).Where("id = ?", user.ID).Updates(user).Error; err != nil {
		handleError(w, "updateUser", fmt.Errorf("error updating user: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
	logUserAction("updateUser", "success", map[string]interface{}{"user": user})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		handleError(w, "deleteUser", fmt.Errorf("invalid input data: %v", err), http.StatusBadRequest)
		return
	}

	if user.ID == 0 {
		handleError(w, "deleteUser", fmt.Errorf("user ID is required"), http.StatusBadRequest)
		return
	}

	if err := db.Delete(&User{}, user.ID).Error; err != nil {
		handleError(w, "deleteUser", fmt.Errorf("error deleting user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
	logUserAction("deleteUser", "success", map[string]interface{}{"id": user.ID})
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			handleError(w, "serveIndex", fmt.Errorf("failed to serve index.html: %v", err), http.StatusInternalServerError)
		}
	}()

	filePath := "index.html"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		handleError(w, "serveIndex", fmt.Errorf("file not found: %s", filePath), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
	logUserAction("serveIndex", "success", map[string]interface{}{"path": r.URL.Path})
}

func main() {
	initLogger()
	initDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/create", createUser)
	mux.HandleFunc("/read", getUsers)
	mux.HandleFunc("/readByID", getUserByID)
	mux.HandleFunc("/update", updateUser)
	mux.HandleFunc("/delete", deleteUser)
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/log-error", logClientError)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./"))))

	logger.Info("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", rateLimiterMiddleware(mux)))
}
