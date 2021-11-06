package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/grpclog"
)

const cartKey = "cart"
const concurrency = 8
const maxRequests = int64(100)
const casAttempts = 100

var etcdHosts = []string{"etcd1:2379",
	"etcd2:2379",
	"etcd3:2379",
	"etcd4:2379",
	"etcd5:2379",
}

func getCart(cli *clientv3.Client, ctx context.Context) *mvccpb.KeyValue {
	resp, err := cli.Get(ctx, cartKey)
	if err != nil {
		log.Fatalf("error getting the key from etcd: %v", err)
	}

	if len(resp.Kvs) != 1 || resp.More {
		log.Fatalf("expected 1 key to be returned but got %v, isMore %v", len(resp.Kvs), resp.More)
	}

	return resp.Kvs[0]
}

func compareAndAdd(cli *clientv3.Client, ctx context.Context, cartItem interface{}) bool {
	getResponse := getCart(cli, ctx)

	currentCartBytes := getResponse.Value
	// deserialize
	var currentCart []interface{}
	err := json.Unmarshal(currentCartBytes, &currentCart)

	if err != nil {
		log.Fatalf("failed to unmarshal etcd value: %v", err)
		return false
	}
	targetCart := append(currentCart, cartItem)
	targetCartBytes, err := json.Marshal(targetCart)
	if err != nil {
		log.Fatalf("failed to marshal etcd value: %v", err)
		return false
	}

	resp, err := cli.Txn(ctx).If(
		clientv3.Compare(clientv3.Value(cartKey), "=", string(currentCartBytes)),
		clientv3.Compare(clientv3.Version(cartKey), "=", getResponse.Version),
	).Then(
		clientv3.OpPut(cartKey, string(targetCartBytes)),
	).Commit()

	if err != nil {
		log.Fatalf("error doing etcd transaction: %v", err)
	}

	return resp.Succeeded
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

func addProductToCart(cli *clientv3.Client, ctx context.Context, requestNumber int64) error {

	cartItem := generateCartItem(requestNumber)

	for i := 0; i < casAttempts; i++ {
		succeeded := compareAndAdd(cli, ctx, cartItem)
		if succeeded {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("failed to add product to cart after %d attempts", casAttempts))
}

func fillCart(cli *clientv3.Client, ctx context.Context, goroutineNumber int64, wg *sync.WaitGroup) {
	log.Printf("filling cart in goroutine number %v", goroutineNumber)
	for i := int64(0); i < maxRequests; i++ {
		err := addProductToCart(cli, ctx, goroutineNumber*maxRequests+i)
		if err != nil {
			log.Printf("Error when adding product to cart %v", err)
		}
	}

	log.Printf("filled cart for goroutine number %v", goroutineNumber)
	wg.Done()
}

func checkState(cli *clientv3.Client, ctx context.Context) {
	log.Print("checking state")
	expectedCartSize := maxRequests * concurrency

	getResponse := getCart(cli, ctx)
	currentCartBytes := getResponse.Value
	// deserialize
	var currentCart []interface{}
	err := json.Unmarshal(currentCartBytes, &currentCart)

	if err != nil {
		log.Fatalf("failed to unmarshal etcd value: %v", err)
	}

	if int64(len(currentCart)) != expectedCartSize {
		log.Println("Unexpected cart size detected")
	} else {
		log.Println("cart has expected size")
	}
}

func main() {
	clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdHosts,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("failed to instanciate etcd client: %v", err)
	}
	defer cli.Close()

	ctx := context.Background()

	log.Print("starting")
	t1 := time.Now()
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := int64(0); i < concurrency; i++ {
		go fillCart(cli, ctx, i+1, &wg)
	}

	log.Print("waiting goroutines")
	wg.Wait()

	t2 := time.Now()
	diff := t2.Sub(t1)
	log.Printf("took %d microseconds", diff.Microseconds())
	// check
	checkState(cli, ctx)
}
