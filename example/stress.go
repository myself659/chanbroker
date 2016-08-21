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
			// fmt.Println(sub, "exit")
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
			// return // 必须退出goroutine，可能出现再次超时，导致重复关闭
		}
	}
}

func RegDo(cb *ChanBroker.ChanBroker, size uint) {
	sub, err := cb.RegSubscriber(size)
	if err != nil {
		fmt.Println(err)
	} else {
		go SubscriberDo(sub, cb)
	}
}

func PubDo(cb *ChanBroker.ChanBroker) {
	var j uint
	for j = 0; j < 100; j++ {
		RegDo(cb, j)
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
			for r := 0; r < 100; r++ {
				go RegDo(cb, uint(r))
			}
			for s := 0; s < 10; s++ {
				go StopDo(cb)
			}
			for r := 0; r < 100; r++ {
				go RegDo(cb, uint(r))
			}
			return
		}

	}

}
func StopDo(cb *ChanBroker.ChanBroker) {
	cb.StopPublish()
}

func main() {
	fmt.Println("start")
	for i := 0; i < 10; i++ {
		cb := ChanBroker.NewChanBroker(time.Second)
		go PubDo(cb)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
