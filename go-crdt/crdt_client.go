package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const concurrency = 8
const maxRequests = int64(1000)
const serverEndpoint = "http://172.17.0.2:12347"
const numServers = 5

type CartItem struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int64   `json:"quantity"`
}

type CartState struct {
	Set          [][2]interface{} `json:"set"` // array of 2-element tuples (first elem is uuid, second elem is cart_item struct)
	TombstoneSet [][2]interface{} `json:"tombstone_set"`
}

func generateSetItem(requestNumber int64) [2]interface{} {
	return [2]interface{}{uuid.New().String(),
		CartItem{
			ProductID:   requestNumber,
			ProductName: fmt.Sprintf("product %d", requestNumber),
			Price:       (rand.Float64() * 5) + 1,
			Quantity:    int64(rand.Intn(5) + 1)}}
}

func sendState(state interface{}, endpoint string) error {
	postBody, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "An Error Occured during marshalling")
	}

	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(endpoint+"/sync_state", "application/json", responseBody)

	if err != nil {
		return errors.Wrap(err, "An Error Occured posting state")
	}
	defer resp.Body.Close()

	// read the body so that connection may be reused
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error while reading sync state response")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Wrapf(err, "Unexpected response code %d while posting state", resp.StatusCode)
	}

	return nil
}

func addProductToCart(requestNumber int64) error {
	// send request to serverEndpoint
	payload := CartState{
		Set:          [][2]interface{}{generateSetItem(requestNumber)},
		TombstoneSet: make([][2]interface{}, 0),
	}

	return sendState(payload, serverEndpoint)
}

func fillCart(goroutineNumber int64, wg *sync.WaitGroup) {
	log.Printf("filling cart in goroutine number %v", goroutineNumber)
	for i := int64(0); i < maxRequests; i++ {
		err := addProductToCart(goroutineNumber*maxRequests + i)
		if err != nil {
			log.Printf("Error when adding product to cart %v", err)
		}
	}

	log.Printf("filled cart for goroutine number %v", goroutineNumber)
	wg.Done()
}

func getState(endpoint string) (interface{}, error) {
	resp, err := http.Get(endpoint + "/raw_state")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed interface{}
	err = json.Unmarshal(body, &parsed)

	if err != nil {
		return nil, err
	}

	return parsed, nil
}

func getDeduplicatedState() ([]interface{}, error) {
	resp, err := http.Get(serverEndpoint + "/state")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed interface{}
	err = json.Unmarshal(body, &parsed)

	if err != nil {
		return nil, err
	}

	return parsed.([]interface{}), nil
}

func checkState() {
	log.Print("checking state")
	expectedCartSize := maxRequests * concurrency
	unexpectedDetected := false

	for i := 0; i < numServers; i++ {
		state, err := getDeduplicatedState()

		if err != nil {
			log.Printf("Error when getting state %v", err)
		}

		if int64(len(state)) != expectedCartSize {
			unexpectedDetected = true
		}
	}

	if unexpectedDetected {
		log.Println("Unexpected cart size detected")
	} else {
		log.Printf("cart has expected size on %d servers (assuming round-robin load-balancing)", numServers)
	}
}

func syncServers() error {

	// under load nginx spawns multiple threads and becomes not exactly round-robbin
	// i.e. subsequent requests may end up on the same server
	// to properly sync the servers we make some direct requests bypassing nginx
	ips := []string{"10.0.3.80",
		"10.0.3.72",
		"10.0.3.160",
		"10.0.3.217",
		"10.0.3.4",
		// second round to sync all the nodes
		"10.0.3.80",
		"10.0.3.72",
		"10.0.3.160",
		"10.0.3.217",
		"10.0.3.4"}
	for i := 0; i < len(ips)-1; i++ {
		state, err := getState("http://" + ips[i] + ":12347")
		if err != nil {
			return err
		}

		err = sendState(state, "http://"+ips[i+1]+":12347")
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	noSyncPtr := flag.Bool("no-sync", false, "Do not sync servers after filling cart. This may be useful when testing asyncronous-replication where we do not care about perfect syncronisation")
	flag.Parse()

	log.Print("starting")
	t1 := time.Now()
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := int64(0); i < concurrency; i++ {
		go fillCart(i+1, &wg)
	}

	log.Print("waiting goroutines")
	wg.Wait()

	if !*noSyncPtr {
		log.Print("syncing servers")
		// sync servers between each other (assuming round-robin balancing is used)
		err := syncServers()
		if err != nil {
			log.Printf("Error when syncing servers %v", err)
		}
	}

	t2 := time.Now()
	diff := t2.Sub(t1)
	log.Printf("took %d microseconds", diff.Microseconds())
	// check
	checkState()
}
