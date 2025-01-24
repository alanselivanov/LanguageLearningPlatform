package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

func TestEmailAndButtonFunctionality(t *testing.T) {

	caps := selenium.Capabilities{"browserName": "MicrosoftEdge"}

	wd, err := selenium.NewRemote(caps, "http://localhost:51492")
	if err != nil {
		t.Fatal(err)
	}
	defer wd.Quit()

	if err := wd.Get("http://localhost:8080"); err != nil {
		t.Fatal(err)
	}

	if err := scrollToBottom(wd); err != nil {
		t.Fatal("Error while scrolling to the bottom of the page:", err)
	}

	startLearningButton, err := wd.FindElement(selenium.ByLinkText, "Start Learning Now")
	if err != nil {
		t.Fatal("Failed to find the button after scrolling:", err)
	}

	_, err = wd.ExecuteScript("arguments[0].scrollIntoView(true); window.scrollBy(0, -100);", []interface{}{startLearningButton})
	if err != nil {
		t.Fatal("Error while scrolling to the button:", err)
	}

	isDisplayed, err := startLearningButton.IsDisplayed()
	if err != nil {
		t.Fatal("Error while checking the visibility of the button:", err)
	}
	if !isDisplayed {
		t.Fatal("The 'Start Learning Now' button is not visible even after scrolling")
	}

	if err := startLearningButton.Click(); err != nil {
		t.Fatal("Error while clicking the button:", err)
	}

	currentURL, err := wd.CurrentURL()
	if err != nil {
		t.Fatal(err)
	}
	if currentURL != "http://localhost:8080/signup_page.html" {
		t.Fatalf("Unexpected URL after clicking the button: %s", currentURL)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("name", "Test User")
	_ = writer.WriteField("email", "testuser@example.com")
	_ = writer.WriteField("message", "This is a test support message.")

	filePath := "./testfile.txt"
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", "testfile.txt")
	if err != nil {
		t.Fatalf("Failed to create file part: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to write file to request: %v", err)
	}

	writer.Close()
	url := "http://localhost:8080/send-support-ticket"

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK but got %d", resp.StatusCode)
	}

	t.Log("Support ticket sent successfully")
}

func scrollToBottom(wd selenium.WebDriver) error {
	for {
		_, err := wd.ExecuteScript("window.scrollBy(0, 500);", nil)
		if err != nil {
			return fmt.Errorf("Error while scrolling the page: %v", err)
		}

		time.Sleep(500 * time.Millisecond)
		return nil
	}
}
