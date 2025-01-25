package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a App

// TestMain is the entry point for all tests in this package.
// It initializes the application, sets up the database, and runs the tests.
func TestMain(m *testing.M) {

	// Initialise the app with database credentials 
	err := a.Initialise(DbUser, DbPassword, "test")
	if err != nil {
		log.Fatal("error occured while initialising the database")
	}

	// Create the required table for testing 
	createTable()

	// Runs all test 
	m.Run()
}

// createTable creates the products table if it does not already exists  
func createTable() {
	createTableQuery := `CREATE TABLE IF NOT EXISTS products(
	id int NOT NULL AUTO_INCREMENT,
	name varchar(255) NOT NULL,
	quantity int,
	price float(10,3),
	PRIMARY KEY(id)
	);`
	_, err := a.DB.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

// clearTable removes all all data from the products table and resets the auto incrementer to 1
func clearTable(){
	a.DB.Query("DELETE from products")
	a.DB.Query("ALTER table products AUTO_INCREMENT=1")
}

// addProduct inserts a product into the products table for testing purposes 
func addProduct(name string, quantity int, price float64) {
	query := fmt.Sprintf("insert into products(name, quantity, price) values('%v', %v, %v)", name, quantity, price)
	_, err := a.DB.Query(query)
	if err != nil {
		log.Println(err)
	}
}

// checkStatusCode compares the expected and actual stautus code and reports an error if they differ
func checkStatusCode(t *testing.T, expectedStatusCode, actualStatusCode int) {
	if expectedStatusCode != actualStatusCode {
		t.Errorf("Expected status: %v, Received %v", expectedStatusCode, actualStatusCode)
	}
}

// sendRequest sends an HTTP request and returns the response recorder for testing.
func sendRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder() 
	a.Router.ServeHTTP(recorder, request)  // Process the request through the app router 
	return recorder
}

// TestGetProduct tests the functionality of the GET /product/1 endpoint.
func TestGetProduct(t *testing.T) {
	clearTable() //Clears the table before the test runs 
	addProduct("keyboard", 400, 3000) // adds a simple product to the products table 

  // Creates a GET request for the newly added product 
	request, _ := http.NewRequest("GET", "/product/1", nil)

  // Sends the request and captures the response 
	response := sendRequest(request)

	// checks that status code is 200 OK 
	checkStatusCode(t, http.StatusOK, response.Code)

	log.Println("GET endpoint Test Successful")
}

// TestCreateTable tests the functionality of the POST  /product endpoint. 
func TestCreateTable(t *testing.T) {
	clearTable() // Clears the table before the test runs 

	// Creates a new product payload 
	var product = []byte(`{"name": "chair", "quantity": 3, "price": 390.00}`)

	//Creates a POST request with the product payload. the payload is converted to bytes buffer 
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(product))
	req.Header.Set("Content-Type", "application/json") // Sets the content type to json 

	// Sends a request and captures the response 
	response := sendRequest(req)

	// Checks that status code is 201 Created. 
	checkStatusCode(t, http.StatusCreated, response.Code)
	
	// Create a map which maps strings to interface and parse the response body into it for validation 
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	// Checks that the response contains the right values. 
	if m["name"] != "chair" {
		t.Errorf("Expected name: %v, Got %v", "chair", m["name"])
	}
	
	if m["quantity"] != 3.0 {
		t.Errorf("Expected quantity: %v, Got %v", 3.0, m["quantity"])
	}

	log.Println("POST endpoint Test Successful")
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProduct("Pen", 5, 900.0)

  // Performs a GET operation on the newly added product and checks the status code 
	request, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(request)
	checkStatusCode(t, http.StatusOK, response.Code)

	// Performs a DELETE operation on the newly added product and checks the status code 
	request, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = sendRequest(request)
	checkStatusCode(t, http.StatusOK, response.Code)

  // Tried to perform another  GET operation and checks the status code 
	request, _ = http.NewRequest("GET", "/product/1", nil)
	response = sendRequest(request)
	checkStatusCode(t, http.StatusNotFound, response.Code)

	log.Println("DELETE endpoint Test Successful")
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProduct("Pen", 5, 900.0)

  // Performs a GET operation on the newly added product and checks the status code 
	request, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(request)
	var oldValue map[string]interface{} // Parses the response body to a map 
	json.Unmarshal(response.Body.Bytes(), &oldValue)

	var product = []byte(`{"name": "Pen", "quantity": 5, "price": 390.00}`) // creates a new product paylaod 

	// Performs  a PUT request with the product payload and converts to bytes buffer 
	req, _ := http.NewRequest("PUT", "/product/1", bytes.NewBuffer(product))
	req.Header.Set("Content-Type", "application/json")
	response = sendRequest(req)
	var newValue map[string]interface{} // Parses the response body to a map 
	json.Unmarshal(response.Body.Bytes(), &newValue)
	
	success := true 

	if oldValue["id"] != newValue["id"]{
		t.Errorf("Expected id: %v, Got %v", newValue["id"], oldValue["id"])
		success = false 
	}

	if oldValue["name"] != newValue["name"]{
		t.Errorf("Expected name: %v, Got %v", newValue["name"], oldValue["name"])
		success = false 
	}
	
	if oldValue["quantity"] != newValue["quantity"]{
		t.Errorf("Expected quantity: %v, Got %v", newValue["quantity"], oldValue["quantity"])
		success = false
	}

	if oldValue["price"] == newValue["price"]{
		t.Errorf("Expected price: %v, Got %v", newValue["price"], oldValue["price"])
		success = false
	}
	
	if success {
		log.Println("PUT endpoint Test Successful")
	}
}
