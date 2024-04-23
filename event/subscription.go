package event

import "fmt"

type Subscription struct {
	SystemName string
	Address    string
	Port       int
	EventDefinition
	Listener
}

type SubscriptionKey string

func (subscription Subscription) SubscriptionKey() SubscriptionKey {
	return SubscriptionKey(fmt.Sprintf("%s-%s-%d-%s", subscription.SystemName, subscription.Address, subscription.Port, subscription.EventDefinition.EventType))
}
