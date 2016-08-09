package ChanBroker

type Content interface{}

type Subscriber chan Content

type ChanBroker struct {
	RegSub      chan Subscriber
	UnRegSub    chan Subscriber
	Contents    chan Content
	Stop        chan bool
	Subscribers map[Subscriber]bool
}

func NewChanBroker() *ChanBroker {
	ChanBroker := new(ChanBroker)
	ChanBroker.RegSub = make(chan Subscriber)
	ChanBroker.UnRegSub = make(chan Subscriber)
	ChanBroker.Contents = make(chan Content)
	ChanBroker.Stop = make(chan bool)
	ChanBroker.Subscribers = make(map[Subscriber]bool)
	ChanBroker.run()

	return ChanBroker
}

func (self *ChanBroker) run() {

	go func() {
		for {
			select {
			case content := <-self.Contents:
				for sub := range self.Subscribers {
					sub <- content
				}
			case sub := <-self.RegSub:
				self.Subscribers[sub] = true
			case sub := <-self.UnRegSub:
				delete(self.Subscribers, sub)
				close(sub)
			case <-self.Stop:
				for sub := range self.Subscribers {
					delete(self.Subscribers, sub)
					close(sub)
				}
			}
		}
	}()
}

func (self *ChanBroker) RegSubscriber() Subscriber {
	sub := make(Subscriber)
	self.RegSub <- sub
	return sub
}

func (self *ChanBroker) UnRegSubscriber(sub Subscriber) {
	self.UnRegSub <- sub
}

func (self *ChanBroker) StopPublish() {
	self.Stop <- true
}

func (self *ChanBroker) PubContent(c Content) {
	self.Contents <- c
}
