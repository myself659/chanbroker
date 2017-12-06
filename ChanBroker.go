package ChanBroker

import (
	"container/list"
	"errors"
	"fmt"
	"time"
)

type Content interface{}

type Subscriber chan Content

type ChanBroker struct {
	RegSub      chan Subscriber
	UnRegSub    chan Subscriber
	Contents    chan Content
	Stop        chan bool
	Subscribers map[Subscriber]*list.List
	timeout     time.Duration
	cachenum    uint
	timerchan   chan time.Time
}

var ErrBrokerExit error = errors.New("ChanBroker exit")
var ErrPublishTimeOut error = errors.New("ChanBroker Pulish Time out")
var ErrRegTimeOut error = errors.New("ChanBroker Reg Time out")
var ErrStopTimeOut error = errors.New("ChanBroker Stop Publish Time out")

func NewChanBroker(timeout time.Duration) *ChanBroker {
	ChanBroker := new(ChanBroker)
	ChanBroker.RegSub = make(chan Subscriber)
	ChanBroker.UnRegSub = make(chan Subscriber)
	ChanBroker.Contents = make(chan Content)
	ChanBroker.Stop = make(chan bool)

	ChanBroker.Subscribers = make(map[Subscriber]*list.List)
	ChanBroker.timeout = timeout
	ChanBroker.cachenum = 0
	ChanBroker.timerchan = nil
	ChanBroker.run()

	return ChanBroker
}

func (self *ChanBroker) onContentPush(content Content) {
	for sub, clist := range self.Subscribers {

		for elem := clist.Front(); elem != nil; elem = elem.Next() {
			select {
			case sub <- elem.Value:
				if self.cachenum > 0 {
					self.cachenum--
				}
			default:
				break // block
			}
		}

		len := clist.Len()
		if len == 0 {
			select {
			case sub <- content:
			default:
				contentList.PushBack(content)
				cachenum++
			}
		} else {
			contentList.PushBack(content)
			cachenum++
		}
	}

	if cachenum > 0 && self.timerchan == nil {
		timer := time.NewTimer(self.timeout)
		self.timerchan = timer.C
	}

}

func (self *ChanBroker) onTimerPush() {
	for sub, clist := range self.Subscribers {

		for elem := clist.Front(); elem != nil; elem = elem.Next() {
			select {
			case sub <- elem.Value:
				if self.cachenum > 0 {
					self.cachenum--
				}
			default:
				break // block
			}
		}
	}

	if self.cachenum > 0 {
		timer := time.NewTimer(self.timeout)
		self.timerchan = timer.C
	} else {
		self.timerchan = nil
	}
}

func (self *ChanBroker) run() {

	go func() { // Broker Goroutine
		for {
			select {
			case content := <-self.Contents:
				onContentPush(content)
			case <-self.timerchan:
				onTimerPush()
			case sub := <-self.RegSub:
				clist := list.New()
				self.Subscribers[sub] = clist

			case sub := <-self.UnRegSub:
				_, ok := self.Subscribers[sub]
				if ok {
					delete(self.Subscribers, sub)
					close(sub)
				}

			case <-self.Stop:
				for sub := range self.Subscribers {
					delete(self.Subscribers, sub)
					close(sub)
				}

				return // exit goroutine
			}
		}
	}()
}

func (self *ChanBroker) RegSubscriber(size uint) (Subscriber, error) {
	sub := make(Subscriber, size)
	select {
	case <-time.After(self.timeout):
		return nil, ErrRegTimeOut
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

func (self *ChanBroker) StopPublish() error {
	select {
	case self.Stop <- true:
		return nil
	case <-time.After(self.timeout):
		return ErrStopTimeOut
	}
}

func (self *ChanBroker) PubContent(c Content) error {
	select {
	case <-time.After(self.timeout):
		return ErrPublishTimeOut
	case self.Contents <- c:
		return nil
	}

}
