package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
)

func getDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func TestUsersAreStoredInDatabase(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer([]byte(`{"email":"jason@mccallister.io","password":"somePassword1!"}`)))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	db := getDB()
	db.AutoMigrate(&user{})
	handler := http.HandlerFunc(usersStore(db))
	user := user{}

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("expected the status code to be %v, got %v instead", http.StatusCreated, status)
	}
	db.Where("email = ?", "jason@mccallister.io").First(&user)
	if user.ID == 0 {
		t.Errorf("expected the user ID to exist, got %v instead", user.ID)
	}
	if user.Password == "somePassword1!" {
		t.Errorf("expected to not find the user by plaintext password, found the user with ID: %v", user.ID)
	}
}

func TestEmailAndPasswordAreRequired(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer([]byte(`{"not":"an email","or":"password"}`)))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersStore(getDB()))

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("expected the status code to be %v, got %v instead", http.StatusUnprocessableEntity, status)
	}
	messages := []string{
		"The email field is required",
		"The email field must be minimum 4 char",
		"The email field must be a valid email address",
		"The password field is required",
		"The password field must be minimum 8 char",
	}
	for _, message := range messages {
		if !strings.Contains(rr.Body.String(), message) {
			t.Errorf("expected the %v validation error to be returned, got this instead\n:%v", message, rr.Body.String())
		}
	}
}

func TestEmptyPostRequestReturnsAnError(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("POST", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersStore(getDB()))

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("expected the status code to be %v, got %v instead", http.StatusUnprocessableEntity, status)
	}
}

func TestRequiresPOSTRequest(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersStore(getDB()))

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("expected the status code to be %v, got %v instead", http.StatusMethodNotAllowed, status)
	}
	if rr.Header().Get("content-type") != "application/json" {
		t.Errorf("expected the content-text to be %v, got %v instead", "application/json", rr.Header().Get("content-type"))
	}
	if !strings.Contains(rr.Body.String(), "method not allowed") {
		t.Errorf("expected the JSON response to contain %v, got %v instead", "method not allowed", rr.Body.String())
	}
}
