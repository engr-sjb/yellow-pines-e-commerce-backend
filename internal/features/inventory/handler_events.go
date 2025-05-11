package inventory

import (
	"context"
	"log"
	"sync"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/eventengine"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/eventengine/event"

	"github.com/google/uuid"
)

// subscriberName is the name of this event handler.
const subscriberName event.SubscriberName = "handler_event.inventory"

type servicer interface {
	createInventory(ctx context.Context, pdID uuid.UUID, stkQty uint) error
}

type HandlerEventsConfig struct {
	DoneCh        <-chan struct{}
	InternalSrvWG *sync.WaitGroup
	EventEngine   eventengine.SubscribeRegisterPublisher
	Service       servicer
	AddressChSize uint16
}

type handlerEvent struct {
	*HandlerEventsConfig
	addressCh chan any
}

func NewEventHandler(
	cfg *HandlerEventsConfig,
) *handlerEvent {
	if cfg.AddressChSize == 0 {
		cfg.AddressChSize = 10
	}

	if cfg.DoneCh == nil || cfg.InternalSrvWG == nil || cfg.EventEngine == nil || cfg.Service == nil {
		log.Fatalf(
			"either 'DoneCh', 'EventEngine', 'InternalSrvWG' or 'Service' is nil in '%s'",
			subscriberName,
		)
	}

	he := &handlerEvent{
		HandlerEventsConfig: cfg,
		addressCh:           make(chan any, cfg.AddressChSize),
	}

	// Register eventsNames the product service will emit
	he.registerServiceEvents()

	he.InternalSrvWG.Add(1)
	go he.listen() // todo: use an errCh to send error to here from listen()
	// if he.errCh != nil {
	// 	log.Panic("error listening to events in", subscriberName)
	// }

	return he
}

func (h *handlerEvent) listen() {
	defer h.InternalSrvWG.Done()

	// subscribe to events
	h.addSubscriptions()

	log.Printf("%s is listening...\n", subscriberName)

	for newEvent := range h.addressCh {
		switch ne := newEvent.(type) {
		case *event.ProductCreatedEvent:
			h.productCreatedEventHandler(ne)

		default:
			log.Printf(
				"received unknown event type: %T\n",
				ne,
			)
		}
	}

	log.Printf("shutting down %s\n", subscriberName)
}

func (h *handlerEvent) productCreatedEventHandler(newEvent *event.ProductCreatedEvent) {
	ctx := context.TODO() // todo: get a proper context

	err := h.Service.createInventory(
		ctx,
		newEvent.ProductID,
		newEvent.StockQuantity,
	)

	if err != nil {
		failedEvent := &event.InventoryCreationFailedEvent{
			ProductID: newEvent.ProductID,
		}

		h.EventEngine.Publish(
			&event.Event{
				Name:    failedEvent.GetEventName(),
				Payload: failedEvent,
			},
		)
	}
}

// registerServiceEvents registers eventsNames that this service will be
// emitting/publishing to for other services to subscribe to.
func (h *handlerEvent) registerServiceEvents() {
	// Register eventsNames the product service will emit
	h.EventEngine.RegisterEvents(
		event.InventoryCreationFailedEventName,
	)
}

// addSubscription iterates over subscribeToEventNames array and subscribes to
// various events with addressCh.
func (h *handlerEvent) addSubscriptions() {
	// subscribeToEventNames is an array of all events this subscriber is
	// wants to Subscribe to.
	subscribeToEventNames := [1]event.EventName{
		event.ProductCreatedEventName,
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
				"error in Subscriber: '%s' \nerror subscribing to events: %v\n",
				subscriberName,
				err,
			)
			// h.errCh <- err
			// return
		}
	}

}
