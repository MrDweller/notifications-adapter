# Notifications adapter
A system that subscribes for events and sends them to a frontend for displaying the events as notifications.
Acting as an adapter or middleware for displaying events to a front end.

## Setup 
Create an `.env` file with the following fields,

```
ADDRESS=<address to run system on>
PORT=<port to run system on>
DOMAIN_ADDRESS=<address that will be registerd to the service registry>
DOMAIN_PORT=<port that will be registerd to the service registry>
SYSTEM_NAME=<system name in service registry>
AUTHENTICATION_INFO=<authentication info>

NOTIFICATION_URL=<The frontends URL, to which events will be forwarded>

SERVICE_REGISTRY_ADDRESS=<address of the service registry>
SERVICE_REGISTRY_PORT=<port of the service registry>

CERT_FILE_PATH=<path to cert .pem file>
KEY_FILE_PATH=<path to key .pem file>
TRUSTSTORE_FILE_PATH=<path to truststore .pem file>

EVENT_HANDLER_IMPLEMENTATION=<event handler implementation, (e.i. "rabbitmq-3.12.12" or other)>
```

Create pem certificate cert and key files, store them in a `certificates` directory. Also add a truststore pem file in the `certificates` directory. 

## Requirements

* **golang 1.22**, later versions should also work.

## Start

Run the command,

```
go run . <event> <evnet>
```

The arguments `<evnet>` states what services the notification adapter should query for and when found subscribe to.
For example if the system is started with `go run . stuck`, the notification adapter will query arrowhead for a `stuck` service, 
and then subscribe to the service using the subscribe funcationallity from [event-handler](https://github.com/MrDweller/event-handler).
