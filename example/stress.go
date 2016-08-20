package main

import (
	_ "container/list"
	"fmt"
	"github.com/myself659/ChanBroker"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type event struct {
	id   int
	info string
}

func UnDoSubscriber(sub ChanBroker.Subscriber, cb *ChanBroker.ChanBroker) {
	n := rand.Int63n(100)
	tn := time.Duration(n)
	<-time.After(tn * time.Second)
	cb.UnRegSubscriber(sub)
}

func SubscriberDo(sub ChanBroker.Subscriber, cb *ChanBroker.ChanBroker) {
	n := 100 + rand.Int63n(10000)
	tn := time.Duration(n)
	for {
		select {
		case <-cb.Stop:
			fmt.Println(sub, "exit")
			return
		case c, ok := <-sub:
			if ok == false {
				return
			}
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
		case <-time.After(tn * time.Millisecond):
			cb.UnRegSubscriber(sub)

		}
	}
}

func PubDo(cb *ChanBroker.ChanBroker) {
	var j uint
	for j = 0; j < 1000; j++ {
		sub := cb.RegSubscriber(j)
		go SubscriberDo(sub, cb)
	}
	n := 10 + rand.Intn(100)

	ticker := time.NewTicker(100 * time.Millisecond)
	var prefix string = "pub_"
	i := 0
	for range ticker.C {
		i++
		if i%3 == 0 {
			var temp string
			temp = prefix + strconv.Itoa(i)
			cb.PubContent(temp)
		} else if i%3 == 1 {
			cb.PubContent(i)
		} else {
			ev := event{i, "event"}

			cb.PubContent(ev)
		}
		if i == n {
			ticker.Stop()

			ncb := ChanBroker.NewChanBroker(time.Second)
			go PubDo(ncb)

			cb.StopPublish()
			return
		}

	}

}

func main() {

	for i := 0; i < 10; i++ {
		cb := ChanBroker.NewChanBroker(time.Second)
		go PubDo(cb)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
