package ChanBroker

import (
	"errors"
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

var errBrokerExit error = errors.New("ChanBroker exit")
var errTimeOut error = errors.New("ChanBroker Time out")

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
				_, ok := self.Subscribers[sub]
				if ok {
					delete(self.Subscribers, sub)
					close(sub)
				}
				self.lock.Unlock()

			case <-self.Stop:
				// close(self.Stop)
				self.lock.Lock()
				for sub := range self.Subscribers {
					delete(self.Subscribers, sub)
					close(sub)
				}
				self.lock.Unlock()

				return // exit goroutine
			}
		}
	}()
}

func (self *ChanBroker) RegSubscriber(size uint) (Subscriber, error) {
	sub := make(Subscriber, size)
	select {
	case <-time.After(self.timeout):
		return nil, errTimeOut
	case self.RegSub <- sub:
		return sub, nil
	}

}

func (self *ChanBroker) UnRegSubscriber(sub Subscriber) {
	select {
	case <-time.After(self.timeout):
		return
	case self.UnRegSub <- sub:
		return
	}

}

func (self *ChanBroker) StopPublish() {
	select {
	case self.Stop <- true:
		return
	case <-time.After(self.timeout):
		return
	}
}

func (self *ChanBroker) PubContent(c Content) error {
	select {
	case <-time.After(self.timeout):
		return errTimeOut
	case self.Contents <- c:
		return nil
	}

}
