package main

import (
	_ "container/list"
	"flag"
	"fmt"
	"github.com/myself659/ChanBroker"
	"math/rand"
	_ "strconv"
	"sync"
	"sync/atomic"
	"time"
)

type event struct {
	id   int
	info string
}

var (
	subnum  *int = flag.Int("n", 1000, "subscribers number")
	pubtime *int = flag.Int("t", 5, "pub duration,default value is 5s")
)
var wg sync.WaitGroup
var sum uint64 = 0

func UnDoSubscriber(sub ChanBroker.Subscriber, cb *ChanBroker.ChanBroker) {
	n := rand.Int63n(100)
	tn := time.Duration(n)
	<-time.After(tn * time.Second)
	cb.UnRegSubscriber(sub)
}

func SubscriberDo(sub ChanBroker.Subscriber, cb *ChanBroker.ChanBroker) {
	i := 0
	for {
		select {
		case c, ok := <-sub:
			if ok == false {
				fmt.Println(sub, "has recv:", i)
				atomic.AddUint64(&sum, uint64(i))
				wg.Done()
				return
			}
			switch t := c.(type) {
			case string:
				fmt.Println(sub, "string:", t)
				i++
			case int:
				fmt.Println(sub, "int:", t)
				t = t
				i++
			case *event:
				fmt.Println(sub, "event:", t.id)
				i++
			case event:
				fmt.Println(sub, "event:", t.id)
				i++
			default:
			}
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
	for j = 0; j < uint(*subnum); j++ {
		RegDo(cb, j)
	}
	tn := time.Duration(*pubtime)
	timer := time.NewTimer(tn * time.Second)
	i := 0
	for {
		select {
		//case <-time.After(tn * time.Second): // 错误用法
		case <-timer.C:
			fmt.Println("stop pub:", i)
			cb.StopPublish()
			//fmt.Println("stop pub:", i++)
			wg.Done()
			return
		default:
			//ev := &event{i, "event"}
			ev := event{i, "event"}
			cb.PubContent(ev)
			i++
		}
	}

}
func StopDo(cb *ChanBroker.ChanBroker) {
	cb.StopPublish()
}

func main() {
	flag.Parse()
	fmt.Println("start")
	wg.Add(1 + *subnum)
	cb := ChanBroker.NewChanBroker(time.Second)
	go PubDo(cb)
	wg.Wait()
	fmt.Println("total:", atomic.LoadUint64(&sum))

}
