package main

import (
	"fmt"
	"github.com/myself659/ChanBroker"
	"math/rand"
	"strconv"
	"time"
)

type event struct {
	id   int
	info string
}

func UnDoSubscriber(sub ChanBroker.Subscriber, b *ChanBroker.ChanBroker) {
	n := rand.Int63n(100)
	tn := time.Duration(n)
	<-time.After(tn * time.Second)
	b.UnRegSubscriber(sub)
}

func SubscriberDo(sub ChanBroker.Subscriber, b *ChanBroker.ChanBroker) {
	for {
		select {
		case c := <-sub:
			switch t := c.(type) {
			case string:
				fmt.Println(sub, "string:", t)
			case int:
				fmt.Println(sub, "int:", t)
				t = t
			case event:
				fmt.Println(sub, "event:", t)
			default:

			}
		}
	}

}
func main() {
	b := ChanBroker.NewChanBroker(time.Second)

	for i := 0; i < 1000; i++ {
		sub, _ := b.RegSubscriber(uint(i))
		if sub != nil {
			go SubscriberDo(sub, b)
			go UnDoSubscriber(sub, b)
		}
	}

	ticker := time.NewTicker(time.Second)

	var prefix string = "pub_"
	i := 0
	for range ticker.C {
		i++
		if i%3 == 0 {
			var temp string
			temp = prefix + strconv.Itoa(i)
			b.PubContent(temp)
		} else if i%3 == 1 {
			b.PubContent(i)
		} else {
			ev := event{i, "event"}
			b.PubContent(ev)
		}
		if i == 100 {
			break
		}
	}
	ticker.Stop()
	b.StopPublish()
	<-time.After(time.Second)
	fmt.Println("exit")
}
