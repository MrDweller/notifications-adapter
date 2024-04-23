package event

import (
	"fmt"
)

type Subscriber struct {
	subscription map[SubscriptionKey]Subscription
}

func NewSubscriber() *Subscriber {
	subscription := map[SubscriptionKey]Subscription{}
	return &Subscriber{
		subscription: subscription,
	}
}

func (subscriber *Subscriber) Subscribe(systemName string, address string, port int, event EventDefinition, metadata map[string]string, output chan<- []byte) error {
	listener := NewRabbitmqListener(address, port, event, metadata)
	subscription := Subscription{
		SystemName:      systemName,
		Address:         address,
		Port:            port,
		EventDefinition: event,
		Listener:        listener,
	}
	_, exists := subscriber.subscription[subscription.SubscriptionKey()]
	if exists {
		return fmt.Errorf("already subscribed to %s event", event.EventType)
	}
	subscriber.subscription[subscription.SubscriptionKey()] = subscription
	return subscriber.subscription[subscription.SubscriptionKey()].Listen(output)
}

func (subscriber *Subscriber) Unsubscribe(systemName string, address string, port int, event EventDefinition) error {
	subscription := Subscription{
		SystemName:      systemName,
		Address:         address,
		Port:            port,
		EventDefinition: event,
	}
	err := subscriber.unsubscribe(subscription.SubscriptionKey())
	if err != nil {
		return err
	}
	return nil
}

func (subscriber *Subscriber) unsubscribe(key SubscriptionKey) error {
	listener, exists := subscriber.subscription[key]
	if !exists {
		return fmt.Errorf("not subscribed to %s event", listener.GetListenerId())
	}
	delete(subscriber.subscription, key)
	return listener.Stop()
}

func (subscriber Subscriber) UnsubscribeAllByEvent(event EventDefinition) error {
	var err error
	err = nil
	for key, subscription := range subscriber.subscription {
		if subscription.EventDefinition == event {
			subErr := subscriber.unsubscribe(key)
			if subErr != nil {
				err = subErr
			}

		}
	}

	return err
}

func (subscriber Subscriber) UnsubscribeAll() error {
	var err error
	err = nil
	for key := range subscriber.subscription {
		subErr := subscriber.unsubscribe(key)
		if subErr != nil {
			err = subErr
		}
	}

	return err
}
