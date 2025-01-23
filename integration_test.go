package main

import (
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
