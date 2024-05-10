package notificationadapter

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	eventsubscriber "github.com/MrDweller/event-handler/subscriber"
	eventhandlertypes "github.com/MrDweller/event-handler/types"
	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/orchestrator-connection/orchestrator"
	"github.com/MrDweller/service-registry-connection/models"

	serviceregistry "github.com/MrDweller/service-registry-connection/service-registry"
)

type NotificationAdapter struct {
	models.SystemDefinition
	ServiceRegistryConnection serviceregistry.ServiceRegistryConnection
	OrchestrationConnection   orchestrator.OrchestratorConnection

	SystemAddress string
	SystemPort    int

	NotificationUrl string

	eventsToSubscribeTo []string
	eventSubscriber     eventsubscriber.EventSubscriber
}

func NewNotificationAdapter(address string, port int, domainAddress string, domainPort int, systemName string, serviceRegistryAddress string, serviceRegistryPort int, notificationUrl string, eventsToSubscribeTo []string) (*NotificationAdapter, error) {
	systemDefinition := models.SystemDefinition{
		Address:    domainAddress,
		Port:       domainPort,
		SystemName: systemName,
	}

	serviceRegistryConnection, err := serviceregistry.NewConnection(serviceregistry.ServiceRegistry{
		Address: serviceRegistryAddress,
		Port:    serviceRegistryPort,
	}, serviceregistry.SERVICE_REGISTRY_ARROWHEAD_4_6_1, models.CertificateInfo{
		CertFilePath: os.Getenv("CERT_FILE_PATH"),
		KeyFilePath:  os.Getenv("KEY_FILE_PATH"),
		Truststore:   os.Getenv("TRUSTSTORE_FILE_PATH"),
	})
	if err != nil {
		return nil, err
	}

	serviceQueryResult, err := serviceRegistryConnection.Query(models.ServiceDefinition{
		ServiceDefinition: "orchestration-service",
	})
	if err != nil {
		return nil, err
	}

	serviceQueryData := serviceQueryResult.ServiceQueryData[0]

	orchestrationConnection, err := orchestrator.NewConnection(orchestrator.Orchestrator{
		Address: serviceQueryData.Provider.Address,
		Port:    serviceQueryData.Provider.Port,
	}, orchestrator.ORCHESTRATION_ARROWHEAD_4_6_1, orchestratormodels.CertificateInfo{
		CertFilePath: os.Getenv("CERT_FILE_PATH"),
		KeyFilePath:  os.Getenv("KEY_FILE_PATH"),
		Truststore:   os.Getenv("TRUSTSTORE_FILE_PATH"),
	})
	if err != nil {
		return nil, err
	}

	eventSubscriber, err := eventsubscriber.EventSubscriberFactory(
		eventhandlertypes.EventHandlerImplementationType(os.Getenv("EVENT_HANDLER_IMPLEMENTATION")),
		domainAddress,
		domainPort,
		systemName,
		serviceRegistryAddress,
		serviceRegistryPort,
		serviceregistry.SERVICE_REGISTRY_ARROWHEAD_4_6_1,
		os.Getenv("CERT_FILE_PATH"),
		os.Getenv("KEY_FILE_PATH"),
		os.Getenv("TRUSTSTORE_FILE_PATH"),
	)
	if err != nil {
		return nil, err
	}

	notificationAdapter := &NotificationAdapter{
		SystemDefinition:          systemDefinition,
		ServiceRegistryConnection: serviceRegistryConnection,
		OrchestrationConnection:   orchestrationConnection,

		SystemAddress: address,
		SystemPort:    port,

		NotificationUrl: notificationUrl,

		eventsToSubscribeTo: eventsToSubscribeTo,
		eventSubscriber:     eventSubscriber,
	}
	if err != nil {
		return nil, err
	}

	return notificationAdapter, nil
}

func (notificationAdapter *NotificationAdapter) ReceiveEvent(event []byte) {

	var workEventDto WorkDTO
	if err := json.Unmarshal(event, &workEventDto); err != nil {
		log.Printf("[!] Error received event with unkown structure: %s\n", event)
		return
	}

	log.Printf("[x] Received %s.\n", workEventDto)
	err := notificationAdapter.HandleEvent(
		WorkDTO{
			EventType: workEventDto.EventType,
			WorkId:    workEventDto.WorkId,
			ProductId: workEventDto.ProductId,
		},
	)
	if err != nil {
		log.Printf("[!] Error during handling of the event: %s\n", err)

	}
}

func (notificationAdapter *NotificationAdapter) StartNotificationAdapter() error {

	for {
		time.Sleep(4 * time.Second)

		for _, event := range notificationAdapter.eventsToSubscribeTo {
			err := notificationAdapter.Subscribe(event)
			if err != nil {
				log.Printf("\t[!] Error during subscription of event %s: %s\n", event, err)

			}
		}

	}

}

func (notificationAdapter *NotificationAdapter) StopNotificationAdapter() error {
	return notificationAdapter.eventSubscriber.UnregisterEventSubscriberSystem()
}

func (notificationAdapter *NotificationAdapter) HandleEvent(event WorkDTO) error {
	notifyEventDTO := NotifyEventDTO{
		ExternalSystemSlug: event.ProductId,
		Type:               event.EventType,
	}
	rawBody, _ := json.Marshal(notifyEventDTO)
	requestBody := bytes.NewBuffer(rawBody)

	response, err := http.Post(notificationAdapter.NotificationUrl, "application/json", requestBody)
	if err != nil {
		return err
	}

	log.Printf("response code from notification of event: %s\n", response.Status)
	log.Printf("response from notification of event: %s\n", response.Body)
	return nil
}

func (notificationAdapter *NotificationAdapter) Subscribe(requestedService string) error {
	log.Printf("[*] Subscribing to %s events.\n", requestedService)
	return notificationAdapter.eventSubscriber.Subscribe(eventhandlertypes.EventType(requestedService), notificationAdapter)

}

func (notificationAdapter *NotificationAdapter) Unsubscribe(requestedService string) error {
	err := notificationAdapter.eventSubscriber.Unsubscribe(eventhandlertypes.EventType(requestedService))
	if err != nil {
		return err
	}
	log.Printf("[*] Unsubscribing from %s events.\n", requestedService)
	return nil
}

type NotifyEventDTO struct {
	ExternalSystemSlug string `json:"externalSystemSlug"`
	Type               string `json:"type"`
}

type WorkDTO struct {
	WorkId string `json:"workId"`

	ProductId string `json:"productId"`
	EventType string `json:"eventType"`
}
