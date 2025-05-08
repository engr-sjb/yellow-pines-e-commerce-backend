package eventengine

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/eventengine/event"
)

type Publisher interface {
	Publish(event *event.Event) error // should take in an event, and add to the events map
}

type Subscriber interface {
	Subscribe(toEventName event.EventName, subscriber *event.Subscriber) error // should add an event if does not exist and add a yourEventListenerAddressCh to that event
}

type RegisterPublisher interface {
	Publisher
	RegisterEvents(eventNames ...event.EventName)
}

type SubscribeRegisterPublisher interface {
	Subscriber
	RegisterPublisher
}

type subscribers struct {
	names      []*event.SubscriberName
	addressChs []chan<- any
}

type EventEngineConfig struct {
	DoneCh        <-chan struct{}
	InternalSrvWG *sync.WaitGroup
}

type eventEngine struct {
	*EventEngineConfig
	wg            sync.WaitGroup
	eventEngineCh chan *event.Event                // This is what the event engine listens to for events being published.
	events        map[event.EventName]*subscribers // This is where all events are kept, and subscribers whom have subscribed to that event. //todo: maybe add a Queue System in each event data
	// events        map[string][]*event.Subscriber // This is where all events are kept, and a slice of Subscribers whom have subscribed to that event. //todo: maybe add a Queue System in each event data
	// ctx               *events.Context // todo: i don't remember why i had ctx here. i know it is important
}

func NewEventEngine(cfg *EventEngineConfig) SubscribeRegisterPublisher {
	if cfg == nil {
		log.Fatalln("'eventEngineConfig' can not be nil")
	}

	if cfg.DoneCh == nil || cfg.InternalSrvWG == nil {
		log.Fatalln("either DoneCh or InternalSrvWG is nil")
	}

	e := &eventEngine{
		EventEngineConfig: cfg,
		events:            make(map[event.EventName]*subscribers, 20),
		eventEngineCh:     make(chan *event.Event, 20),
	}

	e.InternalSrvWG.Add(1)
	go e.listen()

	return e
}

func (e *eventEngine) listen() {
	defer e.InternalSrvWG.Done()

	if e.eventEngineCh == nil {
		log.Fatalln("eventEngineCh is nil")
	}

	log.Println("event engine is listening...")

	// e.publishWorkers() // This should shutdown first before the event engine
	//todo: check to see. i think the pub should close the eventEngineCh so we can drain it

	for { // read until the e.DoneCh is signalled.
		select {
		case <-e.DoneCh:
			e.wg.Wait()
			e.shutdownEventEngineCh()
			log.Println("event engine is shutting down")

			log.Println("draining engineCh")
			for ee := range e.eventEngineCh { //block
				e.broadcaster(ee)
			}

			log.Println("subscribers addressCh are shutting down")
			e.shutdownSubscribersAddressCh()
			return

		case event, isOpened := <-e.eventEngineCh:
			if !isOpened {
				log.Println("eventEngineCh is closed")
				return
			}

			e.broadcaster(event)
		}
	}
}

func (e *eventEngine) broadcaster(event *event.Event) {
	subscribers, exists := e.events[event.Name]
	if !exists {
		log.Printf("\033[35m event %v not found. check your event handler\033[0m",
			event.Name,
		)
		return
	}

	const maxPartitionSize = 4
	partitionSize := (len(subscribers.addressChs) / 2) + 1

	if partitionSize < maxPartitionSize {
		// if an event already exists, find the subscribers to that event and broadcast to each of their addressCh.
		for i, addressCh := range subscribers.addressChs {
			if addressCh == nil {
				log.Printf(
					"subscriber ''%v's'' addressCh is nil. check this event handler to make sure it has been initialized",
					subscribers.names[i],
				)
				continue
			}

			addressCh <- event.Payload
		}
		return
	}

	// if partitionSize > maxPartitionSize, then the code below will run else it
	// return early.
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		for i, addressCh := range subscribers.addressChs[:partitionSize] {
			if addressCh == nil {
				log.Printf(
					"subscriber ''%v's'' addressCh is nil. check this event handler to make sure it has been initialized",
					subscribers.names[i],
				)
				continue
			}
			addressCh <- event.Payload
		}
	}()

	for i, addressCh := range subscribers.addressChs[partitionSize:] {
		// if an event already exists, find the subscribers to that event and broadcast to each of their addressCh.
		if addressCh == nil {
			log.Printf(
				"subscriber ''%v's'' addressCh is nil. check this event handler to make sure it has been initialized",
				subscribers.names[i],
			)
			continue
		}
		addressCh <- event.Payload
	}
}

// RegisterEvents adds all events a publisher can publish to, to the [eventEngine].
//
// IMPORTANT: Register an event before you try to publish or subscribe to it.
func (e *eventEngine) RegisterEvents(eventNames ...event.EventName) {
	for _, eventName := range eventNames {
		if _, exists := e.events[(eventName)]; exists {
			log.Println("event already exists")
			continue
		}

		// e.events[(*eventName)].names = e.events[:len(e.events[*eventName].names)-1]

		// e.events[(eventName)].names = make([]*event.SubscriberName, size)
		// e.events[(eventName)].addressChs = make([]chan<- any, size)

		// e.events[(eventName)] = &subscribers{
		// 	names:      make([]*event.SubscriberName, size),
		// 	addressChs: make([]chan<- any, size),
		// }
		// Todo: I need to find a way to "make()" names and addressChs slices in &subscribers{} for proper memory allocation. Fix it!!!!!!!!!!
		e.events[(eventName)] = &subscribers{}
	}

	log.Println("registering event:", eventNames)
}

func (e *eventEngine) Subscribe(toEventName event.EventName, newSubscriber *event.Subscriber) error {
	if _, ok := e.events[toEventName]; !ok {
		return fmt.Errorf(
			"event '%v' not found. check the service whom is responsible for calling 'eventEngine.RegisterEvents(eventName)' to add an event to the eventEngine and make sure they called it and Registered the eventName or check if you passed the right event name",
			toEventName,
		)
	}

	// e.events[toEventName].names[len(e.events[toEventName].names)-1] = &newSubscriber.Name
	// e.events[toEventName].addressChs[len(e.events[toEventName].addressChs)-1] = newSubscriber.AddressCh

	// Todo: the append is making more memory allocation than necessary. Fix it!!!!!!!!!!
	e.events[toEventName] = &subscribers{
		names:      append(e.events[toEventName].names, &newSubscriber.Name),
		addressChs: append(e.events[toEventName].addressChs, newSubscriber.AddressCh),
	}

	return nil
}

func (e *eventEngine) Publish(event *event.Event) error {
	if _, exists := e.events[event.Name]; !exists {
		return fmt.Errorf(
			"event %v not found. check the service which is to publish the event to make sure they called the 'RegisterEvents()'",
			event.Name,
		)
	}

	// e.wg.Add(1)
	// defer e.wg.Done()
	e.eventEngineCh <- event

	return nil
}

func (e *eventEngine) shutdownSubscribersAddressCh() {
	log.Println("waiting to shut addressChs down")

	for _, subscribers := range e.events {
		for _, addressCh := range subscribers.addressChs {
			if addressCh == nil {
				continue
			}
			close(addressCh)
		}
	}

	log.Println("\033[35m addressChs shutting down\033[0m")
}

func (e *eventEngine) shutdownEventEngineCh() {
	log.Println("waiting to shut event engine down")
	close(e.eventEngineCh)
	log.Println("\033[35m event engine shutting down\033[0m")
}

func (e *eventEngine) publishWorkers() {
	numOfWorkers := max((runtime.NumCPU() / 5), 1)

	for range numOfWorkers {
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()

			// for job := range e.jobsCh {
			// 	log.Println("\033[31m deadlock\033[0m")
			// 	e.eventEngineCh <- job
			// 	log.Println("\033[91m done deadlock\033[0m")
			// }

			for {
				select {
				case <-e.DoneCh:
					return

					// case job := <-e.jobsCh:
					// e.eventEngineCh <- job
				}
			}
		}()
	}
}

func (e *eventEngine) shutdownWorkersJobsCh() {
	log.Println("waiting to shut jobsCh down")
	// close(e.jobsCh)
	log.Println("\033[35m jobsCh shutting down\033[0m")
}
