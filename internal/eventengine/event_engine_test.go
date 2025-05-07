package eventengine

import (
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/eventengine/event"
)

func Test_eventEngine(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lshortfile)

	var err error
	doneCh := make(chan struct{})
	InternalSrvWG := sync.WaitGroup{}

	eventEngine := eventEngine{
		EventEngineConfig: &EventEngineConfig{
			DoneCh:        doneCh,
			InternalSrvWG: &InternalSrvWG,
		},
		events:        make(map[event.EventName]*subscribers, 20),
		eventEngineCh: make(chan *event.Event, 1),
		// jobsCh:        make(chan *event.Event, 2),
	}

	InternalSrvWG.Add(1)
	go eventEngine.listen() // go routine 1

	eventTest := event.Event{
		Name: "test.event.engine.event.name",
	}
	eventEngine.RegisterEvents(eventTest.Name)

	// register a subscriber1 for an event.
	subscriberAddressCh1 := make(chan any, 2)
	err = eventEngine.Subscribe(
		eventTest.Name,
		&event.Subscriber{
			Name:      "test_subscriber_name.1",
			AddressCh: subscriberAddressCh1,
		},
	)
	if err != nil {
		close(subscriberAddressCh1)
		t.Error(err)
		return
	}

	// event handler1
	InternalSrvWG.Add(1)
	go func() {
		defer InternalSrvWG.Done()
		count := 1
		for event := range subscriberAddressCh1 {
			log.Printf(
				"\033[32m reading from subscriber 1: %d.\n%+v\n\033[0m",
				count,
				event,
			)

			count++
		}
		log.Println("\033[98m done reading from subscriber 1 events\033[0m")
	}() //go routine 2

	// register a subscriber2 for an event.
	subscriberAddressCh2 := make(chan any, 2)
	err = eventEngine.Subscribe(
		eventTest.Name,
		&event.Subscriber{
			Name:      "test_subscriber_name.2",
			AddressCh: subscriberAddressCh2,
		},
	)
	if err != nil {
		close(subscriberAddressCh2)
		t.Error(err)
		return
	}

	InternalSrvWG.Add(1)
	go func() {
		defer InternalSrvWG.Done()
		count := 1
		for event := range subscriberAddressCh2 {
			log.Printf(
				"\033[35m reading from subscriber 2: %d.\n%+v\n\033[0m",
				count,
				event,
			)

			count++
		}
		log.Println("done reading from subscriber 2 events")
	}() // go routine 3

	// event publisher || main routine
	for i := range 5 {
		eventEngine.Publish(
			&event.Event{
				Name: eventTest.Name,
				Payload: fmt.Sprintf(
					"test payload: %d",
					i+1,
				),
			},
		)
		log.Println("writing", i+1)
	}
	log.Println("done writing events")

	close(doneCh)
	InternalSrvWG.Wait()
}
