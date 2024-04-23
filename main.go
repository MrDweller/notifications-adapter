package main

import (
	"log"
	"os"
	"strconv"

	notificationadapter "github.com/MrDweller/notificationadapter/notification-adapter"
	models "github.com/MrDweller/service-registry-connection/models"
	"github.com/joho/godotenv"
)

type EventData struct {
	EventType string                  `json:"eventType"`
	Payload   string                  `json:"payload"`
	Source    models.SystemDefinition `json:"source"`
	TimeStamp string                  `json:"timeStamp"`
}

type SubscriberData struct {
	EventType        string                  `json:"eventType"`
	MatchMetaData    bool                    `json:"matchMetaData"`
	NotifyUri        string                  `json:"notifyUri"`
	SubscriberSystem models.SystemDefinition `json:"subscriberSystem"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	domainAddress := os.Getenv("DOMAIN_ADDRESS")
	domainPort, err := strconv.Atoi(os.Getenv("DOMAIN_PORT"))
	if err != nil {
		log.Panic(err)
	}
	systemName := os.Getenv("SYSTEM_NAME")

	address := os.Getenv("ADDRESS")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Panic(err)
	}

	serviceRegistryAddress := os.Getenv("SERVICE_REGISTRY_ADDRESS")
	serviceRegistryPort, err := strconv.Atoi(os.Getenv("SERVICE_REGISTRY_PORT"))
	if err != nil {
		log.Panic(err)
	}

	notificationUrl := os.Getenv("NOTIFICATION_URL")

	if len(os.Args[1:]) <= 0 {
		log.Panicln("Must provide arguments of which events that should be subscribed to: go run main.go <event> <event> ...")

	}
	log.Printf("Events to subscribe for: %v", os.Args[1:])

	notificationAdapter, err := notificationadapter.NewNotificationAdapter(
		address,
		port,
		domainAddress,
		domainPort,
		systemName,
		serviceRegistryAddress,
		serviceRegistryPort,
		notificationUrl,
		os.Args[1:],
	)
	if err != nil {
		log.Panic(err)
	}

	err = notificationAdapter.StartNotificationAdapter()
	if err != nil {
		log.Panic(err)
	}

}
