package main

import (
	"fmt"
	"time"

	"github.com/myself659/chanbroker"
)

func subscriberDo(sub chanbroker.Subscriber, b *chanbroker.Broker, id int) {
	for {
		select {
		case c := <-sub:
			switch t := c.(type) {
			case func():
				t()
			default:
			}
		}
	}

}

func drawCat() {
	fmt.Println("draw a cat!")
}

func drawDog() {
	fmt.Println("draw a dog!")
}

func publisherDo(b *chanbroker.Broker) {

	b.PubContent(drawCat)
	b.PubContent(drawDog)

	<-time.After(3 * time.Second)

	b.StopBroker()
}

func main() {
	// launch broker goroutine
	b := chanbroker.NewBroker(time.Second)

	// register  Subscriber and launch  Subscriber goroutines
	sub1, _ := b.RegSubscriber(1)

	go subscriberDo(sub1, b, 1)

	sub2, _ := b.RegSubscriber(1)

	go subscriberDo(sub2, b, 2)

	// launch Publisher goroutine
	go publisherDo(b)

	// after 3.5s, exit process
	<-time.After(3500 * time.Millisecond)

	fmt.Println("exit")
}
