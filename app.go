package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)


type App struct{
	Router *mux.Router
	DB     *sql.DB
}

// Initialise method creates and Initialise a db connection
func (app *App) Initialise(DbUser, DbPassword, DBName string) error {
	connectionString := fmt.Sprintf("%v:%v@tcp(localhost:3306)/%v", DbUser, DbPassword, DBName)
	var err error
	app.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
	return err
	}
	// creates a new mux router and call the handle routes method
	app.Router = mux.NewRouter().StrictSlash(true)	
	app.handleRoutes()
	return nil 
}

// run method sets up a http listner 
func (app *App) Run(address string) {
	log.Fatal(http.ListenAndServe(address, app.Router))
}

// sendResponse method responsible for sending back response and other information like status code
func sendResponse(w http.ResponseWriter, statusCode int, payload interface{})  {
	// response, _ := json.Marshal(payload)
	response, _ := json.MarshalIndent(payload, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

// sendError responsible for sending eror messages
func sendError(w http.ResponseWriter, statusCode int, err string) {
	error_message := map[string]string{"error": err}
	sendResponse(w, statusCode, error_message)
}

// getProducts method for handing the /products route and gets all products
func (app *App) getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := getProducts(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return 
	}
	sendResponse(w, http.StatusOK, products)
}

// getProduct method used to get a particular product based on id.. handles the /product{id} route 

func (app *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid product ID")
		return
	}
	// initalised a new struct with just one value
	p := product{ID: key }
	// recieves an error from getProduct and pass a db pointer to it..
	err = p.getProduct(app.DB)
	if err != nil {
		// creates a switch statement for the two possible errors 
		switch err {
		case sql.ErrNoRows:
			sendError(w, http.StatusNotFound, "product not found")
		default:
			sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	sendResponse(w, http.StatusOK, p)
}

// function handler for createProduct
// in case of post request we need to decode the data gotten from user using Decoder function 
func (app *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	// used the decoded values to create product in the database
	err = p.createProduct(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusCreated, p)
}

//function handler for updateProduct
func (app *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	// function to get the id of the product that a put needs to be made to
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid product ID")
		return
	}
	// function to receive value from the user as json and decoding it into the struct variable
	var p product
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	// funtion receives an error from the updateProduct method and passes the db pointer to it 
	p.ID = key
	err = p.updateProduct(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, p)
}

func (app *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	// function to get the id of the product that a delete needs to be made to
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid product ID")
		return
	}
	// Initialise a new struct with just one value 
	p := product{ID: key}
	// receives an error from the deleteProduct method and passes the db pointer to it
	err = p.deleteProduct(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{"result": "successfully deleted item"})

}
// note: I call methods relating to our App struct using app.
// handleRoutes method is used to handle all routes and method type 
func (app *App) handleRoutes() {
	app.Router.HandleFunc("/products", app.getProducts).Methods("GET") 
	app.Router.HandleFunc("/product/{id}", app.getProduct).Methods("GET") 
	app.Router.HandleFunc("/product", app.createProduct).Methods("POST") 
	app.Router.HandleFunc("/product/{id}", app.updateProduct).Methods("PUT") 
	app.Router.HandleFunc("/product/{id}", app.deleteProduct).Methods("DELETE") 
}

