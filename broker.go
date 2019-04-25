// Package chanbroker a Broker for goroutine, is simliar to kafka
//
// chanbroker has three types of goroutine: Producer, Consumer(Subscriber), Broker
package chanbroker

import (
	"container/list"
	"errors"
	"time"
)

// Content as interface
type Content interface{}

// Subscriber as channel for Content
type Subscriber chan Content

// chanbroker desc
type Broker struct {
	regSub      chan Subscriber
	unRegSub    chan Subscriber
	contents    chan Content
	stop        chan bool
	subscribers map[Subscriber]*list.List
	timeout     time.Duration
	cachenum    uint
	timerChan   <-chan time.Time
}

// ErrBrokerExit represent  broker goroutine exit
var ErrBrokerExit error = errors.New("chanbroker exit")

// ErrPublishTimeOut represent publish context timeout
var ErrPublishTimeOut error = errors.New("chanbroker Pulish Time out")

// ErrRegTimeOut represent Subscriber registration  timeout
var ErrRegTimeOut error = errors.New("chanbroker Reg Time out")

// ErrStopBrokerTimeOut represent stop broker goroutine  timeout
var ErrStopBrokerTimeOut error = errors.New("chanbroker Stop Broker Time out")

// NewBroker create a  new  broker
func NewBroker(timeout time.Duration) *Broker {
	Broker := new(Broker)
	Broker.regSub = make(chan Subscriber)
	Broker.unRegSub = make(chan Subscriber)
	Broker.contents = make(chan Content, 16)
	Broker.stop = make(chan bool, 1)

	Broker.subscribers = make(map[Subscriber]*list.List)
	Broker.timeout = timeout
	Broker.cachenum = 0
	Broker.timerChan = nil
	Broker.run()

	return Broker
}

func (broker *Broker) onContentPush(content Content) {
	for sub, clist := range broker.subscribers {
		loop := true
		for next := clist.Front(); next != nil && loop == true; {
			cur := next
			next = cur.Next()
			select {
			case sub <- cur.Value:
				if broker.cachenum > 0 {
					broker.cachenum--
				}
				clist.Remove(cur)
			default:
				loop = false
			}
		}

		len := clist.Len()
		if len == 0 {
			select {
			case sub <- content:
			default:
				clist.PushBack(content)
				broker.cachenum++
			}
		} else {
			clist.PushBack(content)
			broker.cachenum++
		}
	}

	if broker.cachenum > 0 && broker.timerChan == nil {
		timer := time.NewTimer(broker.timeout)
		broker.timerChan = timer.C
	}

}

func (broker *Broker) onTimerPush() {
	for sub, clist := range broker.subscribers {
		loop := true
		for next := clist.Front(); next != nil && loop == true; {
			cur := next
			next = cur.Next()
			select {
			case sub <- cur.Value:
				if broker.cachenum > 0 {
					broker.cachenum--
				}
				clist.Remove(cur)
			default:
				loop = false
			}
		}
	}

	if broker.cachenum > 0 {
		timer := time.NewTimer(broker.timeout)
		broker.timerChan = timer.C
	} else {
		broker.timerChan = nil
	}
}

func (broker *Broker) run() {

	go func() { // Broker Goroutine
		for {
			select {
			case content := <-broker.contents:
				broker.onContentPush(content)

			case <-broker.timerChan:
				broker.onTimerPush()

			case sub := <-broker.regSub:
				clist := list.New()
				broker.subscribers[sub] = clist

			case sub := <-broker.unRegSub:
				_, ok := broker.subscribers[sub]
				if ok {
					delete(broker.subscribers, sub)
					close(sub)
				}

			case _, ok := <-broker.stop:
				if ok == true {
					close(broker.stop)
				} else {
					if broker.cachenum == 0 {
						for sub := range broker.subscribers {
							delete(broker.subscribers, sub)
							close(sub)
						}
						return
					}
				}
				broker.onTimerPush()
				for sub, clist := range broker.subscribers {
					if clist.Len() == 0 {
						delete(broker.subscribers, sub)
						close(sub)
					}
				}
			}
		}
	}()
}

// RegSubscriber register subscriber
func (broker *Broker) RegSubscriber(size uint) (Subscriber, error) {
	sub := make(Subscriber, size)

	select {

	case <-time.After(broker.timeout):
		return nil, ErrRegTimeOut

	case broker.regSub <- sub:
		return sub, nil
	}

}

// UnRegSubscriber unregister subscriber
func (broker *Broker) UnRegSubscriber(sub Subscriber) {
	select {
	case <-time.After(broker.timeout):
		return

	case broker.unRegSub <- sub:
		return
	}

}

// StopBroker  stop broker goroutine
func (broker *Broker) StopBroker() error {
	select {
	case broker.stop <- true:
		return nil
	case <-time.After(broker.timeout):
		return ErrStopBrokerTimeOut
	}
}

// PubContent publish content
func (broker *Broker) PubContent(c Content) error {
	select {
	case <-time.After(broker.timeout):
		return ErrPublishTimeOut

	case broker.contents <- c:
		return nil
	}

}
