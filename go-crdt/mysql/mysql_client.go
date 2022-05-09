package main

import (
	"database/sql"
	"time"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

const concurrency = 8
const maxRequests = int64(250)
const updAttempts = 100
const dsnString = "root:testrootpassword@(cart-mariadb)/crdtcart"

func initConnection() *sql.DB {
    db, err := sql.Open("mysql", dsnString)
    if err != nil {
        panic(err)
    }
    // See "Important settings" section.
    db.SetConnMaxLifetime(time.Minute * 3)
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(10)

   return db
}

func initQueries(db *sql.DB) (*sql.Stmt, *sql.Stmt) {
    stmtOut, err := db.Prepare("SELECT current_cart FROM cart WHERE id = ?")
	if err != nil {
		panic(err.Error())
	}

    stmtUpd, err := db.Prepare("UPDATE cart SET current_cart = ? WHERE current_cart = ?")
	if err != nil {
		panic(err.Error())
	}

	return stmtOut, stmtUpd
}

func getCart(stmtOut *sql.Stmt) []byte {
    var resp []byte
    err := stmtOut.QueryRow(1).Scan(&resp)
	if err != nil {
		panic(err.Error())
	}

	return resp
}

func updateCart(stmtUpd *sql.Stmt, targetCart []byte, prevCart []byte) (int64, error) {
    res, err := stmtUpd.Exec(targetCart, prevCart)
    if err != nil {
        return 0, err
    }

    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return 0, err
    }

    return rowsAffected, nil
}

func compareAndAdd(stmtOut *sql.Stmt, stmtUpd *sql.Stmt, cartItem interface{}) bool {
	currentCartBytes := getCart(stmtOut)

	// deserialize
	var currentCart []interface{}
	err := json.Unmarshal(currentCartBytes, &currentCart)

	if err != nil {
		log.Fatalf("failed to unmarshal MySQL value: %v", err)
		return false
	}
	targetCart := append(currentCart, cartItem)
	targetCartBytes, err := json.Marshal(targetCart)
	if err != nil {
		log.Fatalf("failed to marshal MySQL value: %v", err)
		return false
	}

	rowsAffected, err := updateCart(stmtUpd, targetCartBytes, currentCartBytes)

	if err != nil {
		log.Fatalf("error doing MySQL transaction: %v", err)
	}

	return rowsAffected != 0
}

type CartItem struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int64   `json:"quantity"`
}

func generateCartItem(requestNumber int64) interface{} {
	return CartItem{
		ProductID:   requestNumber,
		ProductName: fmt.Sprintf("product %d", requestNumber),
		Price:       (rand.Float64() * 5) + 1,
		Quantity:    int64(rand.Intn(5) + 1)}
}

func addProductToCart(stmtOut *sql.Stmt, stmtUpd *sql.Stmt, requestNumber int64) error {

	cartItem := generateCartItem(requestNumber)

	for i := 0; i < updAttempts; i++ {
		succeeded := compareAndAdd(stmtOut, stmtUpd, cartItem)
		if succeeded {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("failed to add product to cart after %d attempts", updAttempts))
}

func fillCart(stmtOut *sql.Stmt, stmtUpd *sql.Stmt, goroutineNumber int64, wg *sync.WaitGroup) {
	log.Printf("filling cart in goroutine number %v", goroutineNumber)
	for i := int64(0); i < maxRequests; i++ {
		err := addProductToCart(stmtOut, stmtUpd, goroutineNumber*maxRequests+i)
		if err != nil {
			log.Printf("Error when adding product to cart %v", err)
		}
	}

	log.Printf("filled cart for goroutine number %v", goroutineNumber)
	wg.Done()
}

func checkState(stmtOut *sql.Stmt, stmtUpd *sql.Stmt) {
	log.Print("checking state")
	expectedCartSize := maxRequests * concurrency

	currentCartBytes := getCart(stmtOut)
	// deserialize
	var currentCart []interface{}
	err := json.Unmarshal(currentCartBytes, &currentCart)

	if err != nil {
		log.Fatalf("failed to unmarshal MySQL value: %v", err)
	}

	if int64(len(currentCart)) != expectedCartSize {
		log.Println("Unexpected cart size detected")
	} else {
		log.Println("cart has expected size")
	}
}

func main() {
    db := initConnection()
    defer db.Close()

    stmtSelect, stmtUpd := initQueries(db)

	defer stmtSelect.Close()
	defer stmtUpd.Close()

	log.Print("starting")
	t1 := time.Now()
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := int64(0); i < concurrency; i++ {
		go fillCart(stmtSelect, stmtUpd, i+1, &wg)
	}

	log.Print("waiting goroutines")
	wg.Wait()

	t2 := time.Now()
	diff := t2.Sub(t1)
	log.Printf("took %d microseconds", diff.Microseconds())
	// check
	checkState(stmtSelect, stmtUpd)
}
