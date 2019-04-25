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
			case callback:
				t.fc(t.data)
			default:
			}
		}
	}

}

func draw(name string) {
	fmt.Println("draw a", name)
}

func feed(name string) {
	fmt.Println("feed a", name)
}

type callback struct {
	data string
	fc   func(string)
}

func publisherDo(b *chanbroker.Broker) {

	b.PubContent(callback{data: "cat", fc: draw})
	b.PubContent(callback{data: "dog", fc: feed})

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
