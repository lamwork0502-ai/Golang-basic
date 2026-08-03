package main

import (
	"fmt"
	"time"
)

// type Course struct {
// 	Title string
// 	Price int
// }

// func main() {
// 	// 1. add 1 channel
// 	ch := make(chan Course)

// 	// 2. Create goroutine
// 	go func() {
// 		course := Course{Title: "Tips 30", Price: 500}
// 		ch <- course // send data to channel;
// 	}()

// 	c := <-ch
// 	fmt.Printf("Receive Course: Title %s, Pirce %d", c.Title, c.Price)
// }

// pubsub in channel
type Message struct {
	OrderId string
	Title   string
}

func publisher(channel chan<- Message, orders []Message) {

	for _, order := range orders {
		channel <- order
		time.Sleep(time.Second * 1)
	}
	close(channel)
}

func subcribe(channel <-chan Message, userName string) {
	for msg := range channel {
		fmt.Printf("usr %s::Orders:%s:: Title:%s", userName, msg.OrderId, msg.Title)
		time.Sleep(time.Second * 1)
	}
}

func main() {
	// create order channel
	orderChannel := make(chan Message)

	// Simulate order
	orders := []Message{
		{Title: "sach", OrderId: "123"},
	}

	// send order to pub
	go publisher(orderChannel, orders)
	go subcribe(orderChannel, "Lam Nguyen 10")
}
