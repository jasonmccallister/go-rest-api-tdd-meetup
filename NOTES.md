# Testing the Request

## Testing POST request is required

```go
func TestRequiresPOSTRequest(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersStore())

    // Act
	handler.ServeHTTP(rr, req)

    // Assert
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
```

Write the code that makes this pass:

```go
if r.Method != http.MethodPost {
    w.WriteHeader(http.StatusMethodNotAllowed)
    return
}
```

This is not required, but since we are building a JSON API, lets update this test to ensure the content-type is correct:

Add this after the status test:

```go
if rr.Header().Get("content-type") != "application/json" {
    t.Errorf("expected the content-text to be %v, got %v instead", "application/json", rr.Header().Get("content-type"))
}
```

Make it pass by setting the content-type header:

```go
w.Header().Set("Content-type", "application/json")
```

That looks better, but we still don't have any indication what went wrong, this is an API so lets return a JSON response with an error message:

```go
if !strings.Contains(rr.Body.String(), "method not allowed") {
    t.Errorf(
        "expected the JSON response to contain %v, got %v instead",
        "method not allowed",
        rr.Body.String(),
    )
}
```

Make it pass by writing the following literal:

```go
w.Write([]byte(`{"error": "method not allowed"}`))
```

We have tested a few things here, that POST is required, the status code, and JSON error. I think this is a good time to move on.

## Testing Request Validation

Now that the POST request is required, we need to ensure the request has the correct fields. Since this is a users endpoint, we know email and password are required. Let's create a new test.

```go
func TestEmailAndPasswordAreRequired(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("POST", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersStore())

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("expected the status code to be %v, got %v instead", http.StatusUnprocessableEntity, status)
	}
}
```

By default, Go returns a 200 status code, so this test fails immediately.

Since we are sending nothing in the request (an empty body) we need to parse the form and check for errors.

```go
err := r.ParseForm()
if err != nil {
    w.WriteHeader(http.StatusUnprocessableEntity)
    w.Write([]byte(`{"error": "method not allowed"}`))
    return
}
```

You can see there is some duplicate code here, but that is only the "second" time we have used it, so lets not abstract that yet - we have not met our business goals yet.

Now, lets update the test name to represent what we actually handled, an empty POST request.

```go
TestEmptyPostRequestReturnsAnError
```

Now, we need to test non-empty POST requests to ensure there is basic validation. For request validation, I like to the use the following package [thedevsaddam/govalidator](https://github.com/thedevsaddam/govalidator) as it mostly matches the Laravel validation I'm used to.

Lets create a new test

```go
func TestEmailAndPasswordAreRequired(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer([]byte(`{"not":"an email","or":"password"}`)))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersStore())

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("expected the status code to be %v, got %v instead", http.StatusUnprocessableEntity, status)
	}
}
```

Let's add the code to check validation:

```go
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
    w.WriteHeader(http.StatusUnprocessableEntity)
    err := map[string]interface{}{"errors": e}
    json.NewEncoder(w).Encode(err)
    return
}
```

Testing a status code makes the test pass, but what about our contract for consuming our API? What does this tell the consumer?

Let's make sure that the email and password errors are returned in the JSON response, add another assertion to our test

```go
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
```

## Testing the user is created

Now that we have throughly vetted our request, its safe to move onto persisting the user in the database. This is where our ealier discussion about how to build webservers in Go comes in handy.

We need an instance of a database inside of the handler, there are two approaches - I'll give you one guess as to the recommended approach.

1. We can create a new db connection from inside the handler using environment variables
2. Since we have a generic function that returns a `http.HandlerFunc` we can pass in a database to use

The first consideration is how many requests can we handle to the database - what if we become the next Basecamp? Can we handle that many database connections? How do we update our tests?

For maximum portabilty, we need to pass the database into the handler. This will unfortunately break all of our other tests, but that is _why_ we right tests - to make sure it all works!

We are going to use an ORM (GORM in this case) but we could use the generic SQL package if needed. Gorm takes care of migrations and automatically converts into structs.

Let's create the test:

```go
func TestUsersAreStoredInDatabase(t *testing.T) {
	// Arrange
	data := []byte(`{"email":"jason@mccallister.io","password":"somePassword1!"}`)
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	db := getDB()
	db.AutoMigrate(&user{})
	handler := http.HandlerFunc(usersStore(db))

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("expected the status code to be %v, got %v instead", http.StatusCreated, status)
	}
}
```

Our first issue is the status code is 200, not 201 - so lets correct that. The fastest way to "make it green" is to just return the status code.

```go
w.WriteHeader(http.StatusCreated)
```

This makes the test pass, but we did not create a user or check the response.

Add the code to make it work:

```go
// before the handler
type userStoreRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userStoreResponse struct {
	ID uint `json:"id"`
}

// in the handler
body, _ := ioutil.ReadAll(r.Body)
defer r.Body.Close()
json.Unmarshal(body, &req)

newUser := user{
	Email:    req.Email,
	Password: req.Password,
}

// persist the user
db.FirstOrCreate(&user{}, newUser)
db.First(&newUser)

resp := userStoreResponse{
	ID: newUser.ID,
}

data, _ := json.Marshal(resp)
w.WriteHeader(http.StatusCreated)
w.Write(data)
```

Update the test to also pull the user from database using the email.

```go
db.Where("email = ?", "jason@mccallister.io").First(&user)
if user.ID == 0 {
	t.Errorf("expected the user ID to exist, got %v instead", user.ID)
}
```

Since we are dealing with user credentials, we need to hash the password before we store it.

Add a test that ensures we cannot find the user by their plaintext password.

```go
if user.Password == "somePassword1!" {
	t.Errorf("expected to not find the user by plaintext password, found the user with ID: %v", user.ID)
}
```

Lets add the code to make it pass:

```go
hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.MinCost)

newUser := user{
	Email:    req.Email,
	Password: string(hash),
}
```

Thats it, a complete endpoint that handles creating a user!

Before we move on, lets clean things up... this is a lof of code in a single function.
