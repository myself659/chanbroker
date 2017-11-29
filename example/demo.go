package main

import (
	"fmt"
	"time"

	"github.com/myself659/ChanBroker"
)

type event struct {
	id   int
	info string
}

func subscriberDo(sub ChanBroker.Subscriber, b *ChanBroker.ChanBroker, id int) {
	for {
		select {
		case c := <-sub:
			switch t := c.(type) {
			case event:
				fmt.Println("SubscriberId:", id, " event:", t)
			default:
			}
		}
	}

}

func publisherDo(b *ChanBroker.ChanBroker) {
	ticker := time.NewTicker(time.Second)
	i := 0
	for range ticker.C {
		ev := event{i, "event"}
		b.PubContent(ev)
		fmt.Println("Publisher:", ev)
		i++
		if 3 == i {
			break
		}
	}
	ticker.Stop()

	b.StopBroker()
}

func main() {
	// launch broker goroutine
	b := ChanBroker.NewChanBroker("topic-1", time.Second)

	// register  Subscriber and launch  Subscriber goroutine

	sub1, _, _ := b.RegSubscriber(1)

	go subscriberDo(sub1, b, 1)

	sub2, _, _ := b.RegSubscriber(1)

	go subscriberDo(sub2, b, 2)

	// launch Publisher goroutine

	go publisherDo(b)

	// after 3.5s, exit process
	<-time.After(3500 * time.Millisecond)

	fmt.Println("exit")
}
