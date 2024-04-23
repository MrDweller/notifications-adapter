package notificationadapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MrDweller/notificationadapter/event"
	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/orchestrator-connection/orchestrator"
	"github.com/MrDweller/service-registry-connection/models"

	serviceregistry "github.com/MrDweller/service-registry-connection/service-registry"
)

type NotificationAdapter struct {
	models.SystemDefinition
	ServiceRegistryConnection serviceregistry.ServiceRegistryConnection
	OrchestrationConnection   orchestrator.OrchestratorConnection
	*event.Subscriber
	eventChannel chan []byte

	SystemAddress string
	SystemPort    int

	NotificationUrl string

	eventsToSubscribeTo []string
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

	notificationAdapter := &NotificationAdapter{
		SystemDefinition:          systemDefinition,
		ServiceRegistryConnection: serviceRegistryConnection,
		OrchestrationConnection:   orchestrationConnection,
		Subscriber:                event.NewSubscriber(),
		eventChannel:              make(chan []byte),

		SystemAddress: address,
		SystemPort:    port,

		NotificationUrl: notificationUrl,

		eventsToSubscribeTo: eventsToSubscribeTo,
	}
	if err != nil {
		return nil, err
	}

	return notificationAdapter, nil
}

func (notificationAdapter *NotificationAdapter) StartNotificationAdapter() error {
	_, err := notificationAdapter.ServiceRegistryConnection.RegisterSystem(notificationAdapter.SystemDefinition)
	if err != nil {
		return err
	}

	go func(eventsToSubscribeTo []string) {
		for {
			time.Sleep(4 * time.Second)

			for _, event := range eventsToSubscribeTo {
				err := notificationAdapter.Subscribe(event)
				if err != nil {
					log.Printf("\t[!] Error during subscription of event %s: %s\n", event, err)

				}
			}

		}

	}(notificationAdapter.eventsToSubscribeTo)

	for receivedEvent := range notificationAdapter.eventChannel {
		var event event.Event
		if err := json.Unmarshal(receivedEvent, &event); err != nil {
			log.Printf("\t[!] Error received event with unkown structure: %s\n", receivedEvent)
			continue
		}

		log.Printf("\t[x] Received %s.\n", event)
		err := notificationAdapter.HandleEvent(event)
		if err != nil {
			log.Printf("\t[!] Error during handling of the event: %s\n", err)

		}
	}

	return nil
}

func (notificationAdapter *NotificationAdapter) StopNotificationAdapter() error {
	err := notificationAdapter.Subscriber.UnsubscribeAll()
	if err != nil {
		return err
	}

	err = notificationAdapter.ServiceRegistryConnection.UnRegisterSystem(notificationAdapter.SystemDefinition)
	if err != nil {
		return err
	}

	return err
}

func (notificationAdapter *NotificationAdapter) HandleEvent(event event.Event) error {
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
	orchestrationResponse, err := notificationAdapter.OrchestrationConnection.Orchestration(
		requestedService,
		[]string{
			"AMQP-INSECURE-JSON",
		},
		orchestratormodels.SystemDefinition{
			Address:    notificationAdapter.Address,
			Port:       notificationAdapter.Port,
			SystemName: notificationAdapter.SystemName,
		},
		orchestratormodels.AdditionalParametersArrowhead_4_6_1{
			OrchestrationFlags: map[string]bool{
				"overrideStore": true,
			},
		},
	)
	if err != nil {
		return err
	}

	if len(orchestrationResponse.Response) <= 0 {
		return errors.New("found no providers")
	}
	providers := orchestrationResponse.Response

	for _, provider := range providers {
		log.Printf("\t[*] Subscribing to %s events on %s at %s:%d.\n", requestedService, provider.Provider.SystemName, provider.Provider.Address, provider.Provider.Port)
		go func(systemName string, address string, port int, serviceDefinition string, metadata map[string]string) {
			err := notificationAdapter.Subscriber.Subscribe(
				systemName,
				address,
				port,
				event.EventDefinition{
					EventType: serviceDefinition,
				},
				metadata,
				notificationAdapter.eventChannel,
			)

			if err != nil {
				log.Printf("\t[*] Error during subscription: %s\n", err)
				return
			}

		}(provider.Provider.SystemName, provider.Provider.Address, provider.Provider.Port, requestedService, provider.Metadata)

	}

	return nil

}

func (notificationAdapter *NotificationAdapter) Unsubscribe(requestedService string) error {
	err := notificationAdapter.Subscriber.UnsubscribeAllByEvent(
		event.EventDefinition{
			EventType: requestedService,
		},
	)
	if err != nil {
		return err
	}
	log.Printf("\t[*] Unsubscribing from %s events.\n", requestedService)
	return nil
}

type NotifyEventDTO struct {
	ExternalSystemSlug string `json:"externalSystemSlug"`
	Type               string `json:"type"`
}
