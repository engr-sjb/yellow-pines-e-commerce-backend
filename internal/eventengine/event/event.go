package event

type SubscriberName string
type EventName string

type Event struct {
	Name    EventName
	Payload any
}

type Subscriber struct {
	Name      SubscriberName // Name of subscriber
	AddressCh chan<- any     // Where a subscriber is listening for events at.
}
