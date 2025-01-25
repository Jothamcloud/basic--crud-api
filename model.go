package main

import (
	"database/sql"
	"errors"
	"fmt"
)

// the json tags help when we are encoding the data into json
type product struct{
	ID int `json:"id"`
	Name string `json:"name"`
	Quantity int `json:"quantity"`
	Price float64 `json:"price"`
}
// this method receives a db pointer and returns a slice of all the products and errors
// the sql query fetches all the rows in the table 
func getProducts(db *sql.DB) ([]product, error) {
	query := "SELECT id, name, quantity, price from products"
	rows, err := db.Query(query)
	
	if err != nil {
	return nil, err 
	}
	// created an empty slice 
	products := []product{}
	// this function iterates over the row and append the data it receives from this row to our slice 
	// rows.Scan scans it into a struct  
	for rows.Next() {
		var p product 
		err := rows.Scan(&p.ID, &p.Name, &p.Quantity, &p.Price)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// create a get product struct method for the product struct. it takes in out db pointer and returns an error 
// the sql query slects a row from the table when the ID is equal to id
// QueryRow method is used only when there is at most 1 row to return 
func (p *product) getProduct(db *sql.DB) error {
	query := fmt.Sprintf("SELECT name, quantity, price FROM products where id=%v", p.ID)
	row := db.QueryRow(query)
	err := row.Scan(&p.Name, &p.Quantity, &p.Price)
	if err != nil {
		return err 
	}
	return nil  
}

// method recieves a db pointer and returns and error 
// Exec method is used to execute the query 
// the LastInsertId method is used to get the ID of the new inserted row and then we stored it into out struct ID
// you have to specify the fields you are inserting into 
func (p *product) createProduct(db *sql.DB) error {
	query := fmt.Sprintf("insert into products(name, quantity, price) values('%v', %v, %v)", p.Name, p.Quantity, p.Price)
	results, err := db.Exec(query)
	if err != nil {
		return err 
	}
	id, err := results.LastInsertId()
	if err != nil {
		return err 
	}
	p.ID = int(id)
	return nil 
}

// receives a db pointer and returns and error 
// query is used to update the product 
func (p *product) updateProduct(db *sql.DB) error {
	query := fmt.Sprintf("update products set name='%v', quantity=%v, price=%v where id=%v", p.Name, p.Quantity, p.Price, p.ID)
	results, _ := db.Exec(query)
	// logic to handle when the user tries to manipulate a row that does not exist  
	rowsAffected, err := results.RowsAffected()
	if rowsAffected == 0{
		return errors.New("no Such row with that id exists")
	}
	return err
}


// receives a db pointer and returns and error 
// query is used to deletes a row  
func (p *product) deleteProduct(db *sql.DB) error {
	query := fmt.Sprintf("delete from products where id=%v", p.ID)
	results, _ := db.Exec(query)
  // loginc to handle when users try to delete a row that does not exists 
	rowsAffected, err := results.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("no such row with that id exists")
	}
	return err 
}
