
### Introduction 

chanbroker,  a Broker for goroutine, is simliar to kafka 

In chanbroker has three types of goroutine:
- Producer
- Consumer(Subscriber) 
- Broker 


### Usage 

code:

```
package main

import (
    "fmt"
    "github.com/myself659/chanbroker"
    "time"
)

type event struct {
    id   int
    info string
}

func subscriberDo(sub chanbroker.Subscriber, b *chanbroker.Broker, id int) {
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

func publisherDo(b *chanbroker.Broker) {
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
    b := chanbroker.NewBroker(time.Second)

    // register  Subscriber and launch  Subscriber goroutine

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

```


output:
```
Publisher: {0 event}
SubscriberId: 1  event: {0 event}
SubscriberId: 2  event: {0 event}
Publisher: {1 event}
SubscriberId: 2  event: {1 event}
SubscriberId: 1  event: {1 event}
Publisher: {2 event}
SubscriberId: 2  event: {2 event}
SubscriberId: 1  event: {2 event}
exit
```



