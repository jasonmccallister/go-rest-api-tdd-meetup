package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/thedevsaddam/govalidator"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/crypto/bcrypt"
)

// user represents a customer of the application
type user struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	Email     string     `gorm:"type:varchar(100);unique_index" json:"email"`
	Password  string     `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func main() {
	// establish a database connection
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&user{})

	// this creates a duplicate route error
	http.HandleFunc("/users", usersIndex(db))

	http.HandleFunc("/users", usersStore(db))

	http.ListenAndServe(":8080", nil)
}

func usersIndex(db *gorm.DB) http.HandlerFunc {
	type userIndexResponse struct {
		Users []user `json:"users"`
	}
	resp := userIndexResponse{}
	users := []user{}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		db.Find(&users)

		resp.Users = users

		data, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func usersStore(db *gorm.DB) http.HandlerFunc {
	type userStoreRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type userStoreResponse struct {
		ID    uint   `json:"id"`
		Token string `json:"token"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := userStoreRequest{}
		w.Header().Set("Content-type", "application/json")

		r.ParseForm()

		// define the rules
		rules := govalidator.MapData{
			"email":    []string{"required", "min:4", "max:20", "email"},
			"password": []string{"required", "min:8", "max:255"},
		}

		// options for the validator
		opts := govalidator.Options{
			Request: r,
			Data:    &req,
			Rules:   rules,
		}

		// create the validator
		v := govalidator.New(opts)

		// actually validate the request
		e := v.ValidateJSON()
		if len(e) >= 1 {
			err := map[string]interface{}{"errors": e}
			json.NewEncoder(w).Encode(err)
			return
		}

		// read the body from the response
		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		json.Unmarshal(body, &req)

		// convert the request into a user struct
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.MinCost)

		newUser := user{
			Email:    req.Email,
			Password: string(hash),
		}

		// persist the user
		db.FirstOrCreate(&user{}, newUser).First(&newUser)

		resp := userStoreResponse{
			ID:    newUser.ID,
			Token: "something",
		}

		data, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusCreated)
		w.Write(data)
	}
}
