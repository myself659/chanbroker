package ChanBroker

import (
	"container/list"
	"errors"
	"time"
)

// Content Content
type Content interface{}

type contentDesc struct {
	ctt       Content
	id        uint64
	timestamp time.Time
	count     uint32
}

// Subscriber ChanBroker
type Subscriber chan Content

// subscriberDesc
type subscriberDesc struct {
	id        uint64
	timestamp time.Time
	start     uint64
	current   uint64
	status    bool
}

// ChanBroker ChanBroker
type ChanBroker struct {
	regSub           chan Subscriber
	unRegSub         chan Subscriber
	contents         chan Content
	cDescs           map[uint64]*contentDesc
	cidCh            chan uint64
	sidCh            chan uint64
	stop             chan bool
	subscribers      map[Subscriber]*list.List
	subscriberDescs  map[Subscriber]*subscriberDesc
	subscriberDescCh chan *subscriberDesc
	timeout          time.Duration
	cachenum         uint
	timerChan        <-chan time.Time
	name             string
	topic            string
	cid              uint64
}

// ErrBrokerExit ErrBrokerExit
var ErrBrokerExit = errors.New("ChanBroker exit")

// ErrPublishTimeOut ErrPublishTimeOut
var ErrPublishTimeOut = errors.New("ChanBroker Pulish Time out")

// ErrRegTimeOut ErrRegTimeOut
var ErrRegTimeOut = errors.New("ChanBroker Reg Time out")

// ErrStopBrokerTimeOut ErrStopBrokerTimeOut
var ErrStopBrokerTimeOut = errors.New("ChanBroker Stop Broker Time out")

// NewChanBroker NewChanBroker
func NewChanBroker(topic string, timeout time.Duration) *ChanBroker {
	Broker := new(ChanBroker)
	Broker.regSub = make(chan Subscriber)
	Broker.unRegSub = make(chan Subscriber)
	Broker.contents = make(chan Content, 16)
	Broker.cidCh = make(chan uint64)
	Broker.sidCh = make(chan uint64)
	Broker.stop = make(chan bool, 1)
	Broker.cDescs = make(map[uint64]*contentDesc)
	Broker.subscribers = make(map[Subscriber]*list.List)
	Broker.subscriberDescs = make(map[Subscriber]*subscriberDesc)
	Broker.subscriberDescCh = make(chan *subscriberDesc)
	Broker.timeout = timeout
	Broker.cachenum = 0
	Broker.timerChan = nil
	Broker.topic = topic
	Broker.name = topic + "-" + time.Now().String()
	Broker.run()

	return Broker
}

func (broker *ChanBroker) onContentPush(content Content) {
	cid := <-broker.cidCh
	cDesc := new(contentDesc)
	cDesc.ctt = content
	cDesc.id = cid
	cDesc.count = 0
	cDesc.timestamp = time.Now()
	broker.cDescs[cid] = cDesc
	isTimerPush := false
	broker.cid = cid
	for sub, sDesc := range broker.subscriberDescs {
		var i uint64
		for i = sDesc.current; i <= broker.cid; i++ {
			select {
			case sub <- broker.cDescs[i].ctt:
				{
					sDesc.current = sDesc.current + 1
				}
			default:
				{
					isTimerPush = true
					break
				}
			}
		}

	}

	if isTimerPush == true {
		if broker.timerChan == nil {
			timer := time.NewTimer(broker.timeout)
			broker.timerChan = timer.C
		}
	} else {
		if broker.timerChan != nil {
			broker.timerChan = nil
		}
	}

}

func (broker *ChanBroker) onTimerPush() {
	var i uint64
	isTimerPush := false
	for sub, sDesc := range broker.subscriberDescs {
		for i = sDesc.current; i <= broker.cid; i++ {
			select {
			case sub <- broker.cDescs[i].ctt:
				{
					sDesc.current = sDesc.current + 1
				}
			default:
				{
					isTimerPush = true
					break
				}
			}
		}
	}

	if isTimerPush == true {
		if broker.timerChan == nil {
			timer := time.NewTimer(broker.timeout)
			broker.timerChan = timer.C
		}
	} else {
		if broker.timerChan != nil {
			broker.timerChan = nil
		}
	}
}

func (broker *ChanBroker) cidDo() {
	var cid uint64 = 1
	for {
		select {
		case broker.cidCh <- cid:
			{
				cid = cid + 1
			}
		case _, ok := <-broker.stop:
			{
				if ok == true {
					close(broker.stop)
				}
				return // exit goroutine
			}
		}
	}
}

func (broker *ChanBroker) sidDo() {
	var sid uint64 = 1
	for {
		select {
		case broker.sidCh <- sid:
			{
				sid = sid + 1
			}
		case _, ok := <-broker.stop:
			{
				if ok == true {
					close(broker.stop)
				}
				return // exit goroutine
			}

		}

	}
}

func (broker *ChanBroker) brokerDo() {
	for {
		select {
		case content := <-broker.contents:
			broker.onContentPush(content)

		case <-broker.timerChan:
			broker.onTimerPush()

		case sub := <-broker.regSub:
			sid := <-broker.sidCh
			desc := new(subscriberDesc)
			desc.id = sid
			desc.timestamp = time.Now()
			desc.start = broker.cid + 1
			broker.subscriberDescs[sub] = desc
			broker.subscriberDescCh <- desc // sync send

		case sub := <-broker.unRegSub:
			_, ok := broker.subscriberDescs[sub]
			if ok {
				delete(broker.subscriberDescs, sub)
				close(sub)
			}

		case _, ok := <-broker.stop:
			if ok == true {
				close(broker.stop)
			}
			broker.onTimerPush()
			for sub, sDesc := range broker.subscriberDescs {
				if sDesc.current >= broker.cid {
					delete(broker.subscriberDescs, sub)
					close(sub)
				}
			}
		}
	}
}
func (broker *ChanBroker) run() {
	go broker.cidDo()
	go broker.sidDo()
	go broker.brokerDo()
}

// RegSubscriber  RegSubscriber
func (broker *ChanBroker) RegSubscriber(size uint) (Subscriber, uint64, error) {
	sub := make(Subscriber, size)

	select {

	case <-time.After(broker.timeout):
		{
			return nil, 0, ErrRegTimeOut
		}

	case broker.regSub <- sub:
		{
			// get subscriber id
			desc := <-broker.subscriberDescCh
			id := desc.id
			return sub, id, nil
		}
	}

}

// UnRegSubscriber UnRegSubscriber
func (broker *ChanBroker) UnRegSubscriber(sub Subscriber) {
	select {
	case <-time.After(broker.timeout):
		return

	case broker.unRegSub <- sub:
		return
	}

}

// StopBroker StopBroker
func (broker *ChanBroker) StopBroker() error {
	select {
	case broker.stop <- true:
		return nil
	case <-time.After(broker.timeout):
		return ErrStopBrokerTimeOut
	}
}

// PubContent PubContent
func (broker *ChanBroker) PubContent(c Content) error {
	select {
	case <-time.After(broker.timeout):
		return ErrPublishTimeOut

	case broker.contents <- c:
		return nil
	}

}
