package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestGetUserByID(t *testing.T) {
	initLogger()
	InitDB()
	defer Db.Exec("DELETE FROM users")

	testUser := User{
		Name:             "Test User",
		Email:            "testuser@example.com",
		Password:         "password123",
		Role:             "user",
		ConfirmationCode: "12345",
		Confirmed:        true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := Db.Create(&testUser).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	request, err := http.NewRequest("GET", "/readByID?id="+strconv.Itoa(int(testUser.ID)), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	response := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/readByID", getUserByID)
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Incorrect status code. Expected: %d, Got: %d", http.StatusOK, response.Code)
	}

	var returnedUser User
	if err := json.Unmarshal(response.Body.Bytes(), &returnedUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if returnedUser.Name != testUser.Name {
		t.Errorf("Incorrect user name. Expected: %s, Got: %s", testUser.Name, returnedUser.Name)
	}
	if returnedUser.Email != testUser.Email {
		t.Errorf("Incorrect user email. Expected: %s, Got: %s", testUser.Email, returnedUser.Email)
	}
}
func TestCreateUser(t *testing.T) {
	initLogger()
	InitDB()
	defer Db.Exec("DELETE FROM users")

	user := map[string]string{
		"name":     "Test User",
		"email":    "testuser@example.com",
		"password": "password123",
	}

	body, _ := json.Marshal(user)
	request, err := http.NewRequest("POST", "/create", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/create", CreateUser)
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Errorf("Incorrect status code. Expected: %d, Got: %d", http.StatusCreated, response.Code)
	}

	var createdUser User
	if err := json.Unmarshal(response.Body.Bytes(), &createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdUser.Name != user["name"] || createdUser.Email != user["email"] {
		t.Errorf("User data does not match. Expected Name: %s, Got: %s. Expected Email: %s, Got: %s",
			user["name"], createdUser.Name, user["email"], createdUser.Email)
	}
}

func TestUpdateUser(t *testing.T) {
	initLogger()
	InitDB()
	defer Db.Exec("DELETE FROM users")

	testUser := User{
		Name:             "Old Name",
		Email:            "oldemail@example.com",
		Password:         "password123",
		Role:             "user",
		ConfirmationCode: "12345",
		Confirmed:        true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := Db.Create(&testUser).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	updatedData := map[string]interface{}{
		"id":    testUser.ID,
		"name":  "Updated Name",
		"email": "updatedemail@example.com",
	}
	body, _ := json.Marshal(updatedData)
	request, err := http.NewRequest("PUT", "/update", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/update", updateUser)
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Incorrect status code. Expected: %d, Got: %d", http.StatusOK, response.Code)
	}

	var updatedUser User
	if err := json.Unmarshal(response.Body.Bytes(), &updatedUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updatedUser.Name != updatedData["name"] || updatedUser.Email != updatedData["email"] {
		t.Errorf("User data does not match. Expected Name: %s, Got: %s. Expected Email: %s, Got: %s",
			updatedData["name"], updatedUser.Name, updatedData["email"], updatedUser.Email)
	}
}

func TestDeleteUser(t *testing.T) {
	initLogger()
	InitDB()
	defer Db.Exec("DELETE FROM users")

	testUser := User{
		Name:             "Test User",
		Email:            "testuser@example.com",
		Password:         "password123",
		Role:             "user",
		ConfirmationCode: "12345",
		Confirmed:        true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := Db.Create(&testUser).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	body, _ := json.Marshal(map[string]uint{"id": testUser.ID})
	request, err := http.NewRequest("DELETE", "/delete", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/delete", deleteUser)
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Incorrect status code. Expected: %d, Got: %d", http.StatusOK, response.Code)
	}

	var deletedUser User
	if err := Db.First(&deletedUser, testUser.ID).Error; err == nil {
		t.Errorf("User was not deleted from database. User ID: %d", testUser.ID)
	}
}
