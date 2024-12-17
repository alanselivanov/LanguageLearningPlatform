package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Request struct {
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/post", handlePostRequest)
	http.HandleFunc("/get", handleGetRequest)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Only POST requests are allowed")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendErrorResponse(w, "Unable to read request body")
		return
	}

	var requestData Request
	err = json.Unmarshal(body, &requestData)
	if err != nil || requestData.Message == "" {
		sendErrorResponse(w, "Invalid JSON message")
		return
	}

	fmt.Println("received data:", requestData.Message)

	response := Response{
		Status:  "success",
		Message: "Data successfully received",
	}
	sendJSONResponse(w, response)
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Only GET requests are allowed")
		return
	}

	response := Response{
		Status:  "success",
		Message: "GET request successfully received",
	}
	sendJSONResponse(w, response)
}

func sendJSONResponse(w http.ResponseWriter, response Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, errorMessage string) {
	response := Response{
		Status:  "fail",
		Message: errorMessage,
	}
	sendJSONResponse(w, response)
}
