package main

const concurrency = 8
const serverEndpoint = "http://127.0.0.1:12347"

func addProductToCart() {
	// send request to serverEndpoint
}

func fillCart() {
	for i := 0; i < 1000; i++ {
		addProductToCart()
	}
}

func main() {
	for i := 0; i < concurrency; i++ {
		go fillCart()
	}
}
