package main

import (
	"fmt"
	"github.com/myself659/ChanBroker"
	"strconv"
	"time"
)

type event struct {
	id   int
	info string
}

func SubDone(sub2 ChanBroker.Subscriber, b *ChanBroker.ChanBroker) {
	for c := range sub2 {
		switch t := c.(type) {
		case string:
			fmt.Println(sub2, "string:", t)
		case int:
			fmt.Println(sub2, "int:", t)
			t = t
		case event:
			//fmt.Println(sub2, "event:", t)
		default:

		}
	}

}
func main() {
	b := ChanBroker.NewChanBroker(time.Second)

	for i := 0; i < 1000; i++ {
		sub := b.RegSubscriber()
		go SubDone(sub, b)
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
	fmt.Println("exit")
}
