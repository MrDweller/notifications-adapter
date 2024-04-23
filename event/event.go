package event

type EventDefinition struct {
	EventType string `json:"eventType"`
}

type Event struct {
	EventDefinition
	WorkId    string `json:"workId"`
	ProductId string `json:"productId"`
}
