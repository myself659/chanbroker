package ChanBroker

import (
	"fmt"
	"sync"
	"time"
)

type Content interface{}

type Subscriber chan Content

type ChanBroker struct {
	RegSub      chan Subscriber
	UnRegSub    chan Subscriber
	Contents    chan Content
	Stop        chan bool
	exit        bool
	Subscribers map[Subscriber]bool
	lock        sync.RWMutex
	timeout     time.Duration
}

func NewChanBroker(timeout time.Duration) *ChanBroker {
	ChanBroker := new(ChanBroker)
	ChanBroker.RegSub = make(chan Subscriber)
	ChanBroker.UnRegSub = make(chan Subscriber)
	ChanBroker.Contents = make(chan Content)
	ChanBroker.Stop = make(chan bool)
	ChanBroker.exit = false
	ChanBroker.Subscribers = make(map[Subscriber]bool)
	ChanBroker.timeout = timeout
	ChanBroker.run()

	return ChanBroker
}

func (self *ChanBroker) run() {

	go func() {
		for {
			select {
			case content := <-self.Contents:
				go func() {
					self.lock.RLock()
					for sub := range self.Subscribers {
						select {
						case sub <- content:
						case <-time.After(self.timeout):
							fmt.Println(sub, "time out ")
						}

					}
					self.lock.RUnlock()
				}()
			case sub := <-self.RegSub:
				self.lock.Lock()
				self.Subscribers[sub] = true
				self.lock.Unlock()
			case sub := <-self.UnRegSub:
				self.lock.Lock()
				delete(self.Subscribers, sub)
				self.lock.Unlock()
				close(sub)
			case <-self.Stop:
				if self.exit == false {
					self.exit = true
					close(self.Stop)
					self.lock.Lock()
					for sub := range self.Subscribers {
						delete(self.Subscribers, sub)
						close(sub)
					}
					self.lock.Unlock()

					return
				}
			}
		}
	}()
}

func (self *ChanBroker) RegSubscriber(size uint) Subscriber {
	if self.exit == true {
		return nil
	}
	sub := make(Subscriber, size) // maybe block
	self.RegSub <- sub
	return sub
}

func (self *ChanBroker) UnRegSubscriber(sub Subscriber) {
	if self.exit == true {
		return
	}
	self.UnRegSub <- sub // maybe block
}

func (self *ChanBroker) StopPublish() {
	self.Stop <- true // maybe block
}

func (self *ChanBroker) PubContent(c Content) {
	self.Contents <- c // maybe block
}
