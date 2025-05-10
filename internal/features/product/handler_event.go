package product

import (
	"context"
	"log"
	"sync"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/eventengine"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/eventengine/event"
)

// type servicerEvent interface {
// 	deleteProduct(ctx context.Context, productID uuid.UUID) error
// }

// subscriberName is the name of this event handler.
const subscriberName event.SubscriberName = "handler_event.product"

type HandlerEventsConfig struct {
	DoneCh        <-chan struct{}
	InternalSrvWG *sync.WaitGroup
	EventEngine   eventengine.SubscribeRegisterPublisher
	Service       servicer
	AddressChSize uint16
}

type handlerEvents struct {
	*HandlerEventsConfig
	addressCh chan any
}

func NewHandlerEvents(
	cfg *HandlerEventsConfig,
) *handlerEvents {
	if cfg.AddressChSize == 0 {
		cfg.AddressChSize = 10
	}

	if cfg.DoneCh == nil || cfg.EventEngine == nil || cfg.Service == nil {
		log.Fatalf(
			"either 'DoneCh', 'EventEngine' or 'Service' is nil in '%s'",
			subscriberName,
		)
	}

	he := &handlerEvents{
		HandlerEventsConfig: cfg,
		addressCh:           make(chan any, cfg.AddressChSize),
	}

	// Register eventsNames the product service will emit
	he.registerServiceEvents()

	he.InternalSrvWG.Add(1)
	go he.listen()

	return he
}

func (h *handlerEvents) listen() {
	defer h.InternalSrvWG.Done()

	// subscribe to events
	h.addSubscription()

	log.Printf("%s is listening...\n", subscriberName)

	// a for select statement is not used here because the event engine will
	// close the addressCh
	for newEvent := range h.addressCh {
		switch ne := newEvent.(type) {
		case *event.InventoryCreationFailedEvent:
			h.inventoryCreationFailedEventHandler(ne)

		default:
			log.Printf(
				"received unknown event type: %T\n",
				ne,
			)
		}
	}

	log.Printf("shutting down %s\n", subscriberName)
}

func (h *handlerEvents) inventoryCreationFailedEventHandler(
	event *event.InventoryCreationFailedEvent,
) {
	ctx := context.TODO() // todo: get a proper context

	err := h.Service.deleteProduct(
		ctx,
		event.ProductID,
	)
	if err != nil {
		//TODO: push err to Notification service to then push to user via webhook to the client.
		log.Println("error handling deleting")
	}
}

// registerServiceEvents registers eventsNames that this service will be
// emitting/publishing to for other services to subscribe to.
func (h *handlerEvents) registerServiceEvents() {
	// Register eventsNames the product service will emit
	h.EventEngine.RegisterEvents(
		event.ProductCreatedEventName,
	)
}

// addSubscription iterates over subscribeToEventNames array and subscribes to
// various events with addressCh.
func (h *handlerEvents) addSubscription() {
	// subscribeToEventNames is an array of all events this subscriber is
	// wants to Subscribe to.
	subscribeToEventNames := [1]event.EventName{
		event.InventoryCreationFailedEventName,
	}

	// Subscribe to events from the [subscriptions] array. If you want to add
	//  more subscriptions, add it to the [subscribeToEventNames] array.
	var err error
	for _, v := range subscribeToEventNames {
		err = h.EventEngine.Subscribe(
			v,
			&event.Subscriber{
				Name:      subscriberName,
				AddressCh: h.addressCh,
			},
		)
		if err != nil {
			log.Fatalf(
				"error subscribing to events in:%s\nerror subscribing to events: %v\n",
				subscriberName,
				err,
			)
		}
	}
}
