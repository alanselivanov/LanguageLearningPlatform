package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	Password         string    `json:"password"`
	Role             string    `json:"role"`
	ConfirmationCode string    `json:"confirmation_code"`
	Confirmed        bool      `json:"confirmed"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Product struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Price           float64   `json:"price"`
	Characteristics string    `json:"characteristics"`
	Date            time.Time `json:"date"`
	Image           string    `json:"image"`
}

var (
	jwtKey  []byte
	Db      *gorm.DB
	logger  *logrus.Logger
	limiter = rate.NewLimiter(30, 60)
)

func initLogger() {

	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("can't open the file for logs: %v\n", err)
		os.Exit(1)
	}

	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(logFile)
	logger.SetLevel(logrus.DebugLevel)

	logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
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

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}
	jwtKey = []byte(os.Getenv("JWT_SECRET"))
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to the database:", err)
	}

	err = Db.AutoMigrate(&User{}, &Product{})
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

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		handleError(w, "createUser", fmt.Errorf("invalid input data: %v", err), http.StatusBadRequest)
		return
	}

	logger.WithFields(logrus.Fields{
		"name":     user.Name,
		"email":    user.Email,
		"password": user.Password,
		"role":     user.Role,
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
	if user.Role == "" {
		user.Role = "user"
	}

	user.ConfirmationCode = generateConfirmationCode()
	user.Confirmed = false
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := Db.Create(&user).Error; err != nil {
		handleError(w, "createUser", fmt.Errorf("error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	if err := sendConfirmationEmail(user); err != nil {
		handleError(w, "createUser", fmt.Errorf("failed to send confirmation email: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
	logUserAction("createUser", "success", map[string]interface{}{"user_id": user.ID})
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	pageStr := r.URL.Query().Get("page")
	limit := 10
	page := 1

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	offset := (page - 1) * limit

	if err := Db.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		handleError(w, "getUsers", fmt.Errorf("error retrieving users: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
	logUserAction("getUsers", "success", map[string]interface{}{"page": page, "count": len(users)})
}

func getUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		handleError(w, "getUserByID", fmt.Errorf("id is required"), http.StatusBadRequest)
		return
	}

	var user User
	if err := Db.First(&user, id).Error; err != nil {

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
func getUserByIDProf(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var user User
	if err := Db.Where("id = ?", id).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
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
	if user.Role != "" && (user.Role != "user" && user.Role != "admin") {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}
	user.UpdatedAt = time.Now()

	if err := Db.Model(&User{}).Where("id = ?", user.ID).Updates(user).Error; err != nil {
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

	if err := Db.Delete(&User{}, user.ID).Error; err != nil {
		handleError(w, "deleteUser", fmt.Errorf("error deleting user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
	logUserAction("deleteUser", "success", map[string]interface{}{"id": user.ID})
}

func generateJWT(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   user.ID,
		"name": user.Name,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtKey)
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
	if err := Db.Where("name = ?", loginData.Name).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !user.Confirmed {
		http.Error(w, "Account not confirmed. Please check your email.", http.StatusForbidden)
		return
	}

	if user.Password != loginData.Password {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Login successful",
		"token":   token,
		"name":    user.Name,
		"role":    user.Role,
		"id":      user.ID,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		role, ok := claims["role"].(string)
		if !ok || role != "admin" {
			http.Error(w, "Access denied: Admins only", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func generateConfirmationCode() string {
	return strconv.Itoa(time.Now().Nanosecond())
}

func sendConfirmationEmail(user User) error {
	subject := "Подтверждение регистрации"
	body := fmt.Sprintf("Здравствуйте, %s!\n\nПожалуйста, подтвердите вашу регистрацию, перейдя по ссылке: http://localhost:8080/confirm?code=%s", user.Name, user.ConfirmationCode)

	return sendEmail(subject, body, []string{user.Email}, nil)
}

func confirmEmail(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		handleError(w, "confirmEmail", errors.New("confirmation code is required"), http.StatusBadRequest)
		return
	}

	var user User
	if err := Db.Where("confirmation_code = ?", code).First(&user).Error; err != nil {
		handleError(w, "confirmEmail", fmt.Errorf("invalid confirmation code: %v", err), http.StatusNotFound)
		return
	}

	user.Confirmed = true
	user.ConfirmationCode = ""
	if err := Db.Save(&user).Error; err != nil {
		handleError(w, "confirmEmail", fmt.Errorf("error confirming email: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Email confirmed successfully"})
	logUserAction("confirmEmail", "success", map[string]interface{}{"user_id": user.ID})
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

	if err := sendEmailToSupport(subject, body, file, fileHeader); err != nil {
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ticket submitted successfully!")
}

func sendEmailToSupport(subject, body string, attachment io.Reader, fileHeader *multipart.FileHeader) error {
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

	boundary := "boundary-example"
	msg.WriteString(fmt.Sprintf("From: %s\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\n", subject))
	msg.WriteString("MIME-Version: 1.0\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))
	msg.WriteString("\n")

	msg.WriteString(fmt.Sprintf("--%s\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=utf-8\n\n")
	msg.WriteString(body + "\n\n")

	if attachment != nil && fileHeader != nil {
		fileContent, err := io.ReadAll(attachment)
		if err != nil {
			return fmt.Errorf("failed to read file content: %v", err)
		}

		encoded := base64.StdEncoding.EncodeToString(fileContent)

		msg.WriteString(fmt.Sprintf("--%s\n", boundary))
		msg.WriteString(fmt.Sprintf("Content-Type: %s\n", fileHeader.Header.Get("Content-Type")))
		msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\n", fileHeader.Filename))
		msg.WriteString("Content-Transfer-Encoding: base64\n\n")
		msg.WriteString(encoded)
		msg.WriteString("\n\n")
	}

	msg.WriteString(fmt.Sprintf("--%s--", boundary))

	err := smtp.SendMail(smtpHost+":"+smtpPort, smtp.PlainAuth("", smtpUser, smtpPass, smtpHost), from, to, msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func sendEmail(subject, body string, to []string, cc []string) error {
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	headers := make(map[string]string)
	headers["From"] = smtpUser
	headers["To"] = to[0]
	headers["Subject"] = subject

	if len(cc) > 0 {
		headers["Cc"] = cc[0]
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	return smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpHost, smtpPort),
		auth,
		smtpUser,
		to,
		[]byte(message),
	)
}
func createProduct(w http.ResponseWriter, r *http.Request) {
	var product struct {
		Name            string  `json:"name"`
		Description     string  `json:"description"`
		Price           float64 `json:"price"`
		Characteristics string  `json:"characteristics"`
		Date            string  `json:"date"`
		Image           string  `json:"image"`
	}

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse("2006-01-02", product.Date)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD.", http.StatusBadRequest)
		return
	}

	newProduct := Product{
		Name:            product.Name,
		Description:     product.Description,
		Price:           product.Price,
		Characteristics: product.Characteristics,
		Date:            parsedDate,
		Image:           product.Image,
	}

	if err := Db.Create(&newProduct).Error; err != nil {
		http.Error(w, "Failed to save product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProduct)
}

func filterUsers(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")

	var users []User
	query := Db.Model(&User{})

	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if email != "" {
		query = query.Where("email ILIKE ?", "%"+email+"%")
	}

	if err := query.Find(&users).Error; err != nil {
		handleError(w, "filterUsers", fmt.Errorf("error filtering users: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
	logUserAction("filterUsers", "success", map[string]interface{}{
		"filters": map[string]string{
			"name":  name,
			"email": email,
		},
		"count": len(users),
	})
}

func sortUsers(w http.ResponseWriter, r *http.Request) {
	sortField := r.URL.Query().Get("field")
	sortOrder := r.URL.Query().Get("order")

	if sortField == "" {
		sortField = "id"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	var users []User
	query := Db.Order(fmt.Sprintf("%s %s", sortField, sortOrder))

	if err := query.Find(&users).Error; err != nil {
		http.Error(w, fmt.Sprintf("Error sorting users: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
func loginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "login_page.html")
}
func signupPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "signup_page.html")
}
func mainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "main_page.html")
}
func adminPanel(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "adminPanel.html")
}
func profilePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "profile_page.html")
}
func main() {
	initLogger()
	InitDB()
	mux := http.NewServeMux()

	mux.HandleFunc("/confirm", confirmEmail)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/create", CreateUser)
	mux.Handle("/read", adminMiddleware(http.HandlerFunc(getUsers)))
	mux.Handle("/readByID", adminMiddleware(http.HandlerFunc(getUserByID)))
	mux.Handle("/readByIDprof", authMiddleware(http.HandlerFunc(getUserByIDProf)))
	mux.Handle("/update", authMiddleware(http.HandlerFunc(updateUser)))
	mux.Handle("/delete", adminMiddleware(http.HandlerFunc(deleteUser)))
	mux.Handle("/log-error", adminMiddleware(http.HandlerFunc(logClientError)))
	mux.Handle("/send-support-ticket", authMiddleware(http.HandlerFunc(sendSupportTicket)))
	mux.Handle("/filter", adminMiddleware(http.HandlerFunc(filterUsers)))
	mux.Handle("/sort", adminMiddleware(http.HandlerFunc(sortUsers)))
	mux.Handle("/create-product", adminMiddleware(http.HandlerFunc(createProduct)))
	mux.HandleFunc("/static/loginPage", loginPage)
	mux.HandleFunc("/static/signupPage", signupPage)
	mux.HandleFunc("/adminPanel", adminPanel)
	mux.HandleFunc("/profilePage", profilePage)
	mux.HandleFunc("/", mainPage)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	logger.Info("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", rateLimiterMiddleware(mux)))

}
