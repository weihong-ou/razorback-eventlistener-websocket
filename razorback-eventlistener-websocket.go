package main


import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/messaging"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/client_event_pb2"
	"protobuf/events_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/validator_pb2"
	zmq "github.com/pebbe/zmq4"
	"os"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"log"
	"net/http"
)

const (
	DEFAULT_VALIDATOR_URL = "tcp://localhost:4004"
	// Calculated from the 1st 6 characters of SHA-512 - TBD?:
	RAZORBACK_TP_ADDRESS_PREFIX = "a4d219"
)

// Global variable for remembering validator URL
var validatorToConnet = DEFAULT_VALIDATOR_URL

// Get URL from environment if present
func init() {
	validatorUrl := os.Getenv("VALIDATOR_URL")
	if validatorUrl != "" {
		validatorToConnet = validatorUrl
	}
}

func listenToEvents(filters []*events_pb2.EventFilter) error {

	// Create Socket Server
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Println("Connected")
	})

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)

	go func(serveMux *http.ServeMux) {
		log.Println("Starting server...")
		log.Panic(http.ListenAndServe(":3811", serveMux))
	}(serveMux)

	// Listen to cookiejar state-delta events.
	// Create a connection with validator for that
	zmqType := zmq.DEALER
	zmqContext, err := zmq.NewContext()
	if err != nil {
		return err
	}

	zmqConnection, err := messaging.NewConnection(zmqContext, zmqType, validatorToConnet, false)
	// Remember to close the connection when either not needed or error occurs
	if err != nil {
		return err
	}
	defer zmqConnection.Close()

	// Subscribe to events
	blockCommitSubscription := events_pb2.EventSubscription{
		EventType: "sawtooth/block-commit",
	}
	stateDeltaSubscription := events_pb2.EventSubscription{
		EventType: "sawtooth/state-delta",
		Filters:   	filters,
	}
	request := client_event_pb2.ClientEventsSubscribeRequest{
		Subscriptions: []*events_pb2.EventSubscription{
			&blockCommitSubscription,
			&stateDeltaSubscription,
		},
	}
	serializedRequest, err := proto.Marshal(&request)
	if err != nil {
		return err
	}

	// Send the subscription request
	corrId, err := zmqConnection.SendNewMsg(
		validator_pb2.Message_CLIENT_EVENTS_SUBSCRIBE_REQUEST,
		serializedRequest,
	)
	if err != nil {
		return err
	}
	// Wait for subscription status
	_, response, err := zmqConnection.RecvMsgWithId(corrId)
	if err != nil {
		return err
	}
	eventSubscribeResponse := client_event_pb2.ClientEventsSubscribeResponse{}
	err = proto.Unmarshal(response.Content, &eventSubscribeResponse)
	if err != nil {
		return err
	}
	if eventSubscribeResponse.Status !=
		client_event_pb2.ClientEventsSubscribeResponse_OK {
		return errors.New("Client couldn't subscribe successfully")
	}

	// Listen for events in an infinite loop
	println("Listening to events.")
	for {
		_, message, err := zmqConnection.RecvMsg()
		if err != nil {
			return err
		}
		if message.MessageType != validator_pb2.Message_CLIENT_EVENTS {
			return errors.New("Unexpected message received, something which is not subscribed")
		}
		eventList := events_pb2.EventList{}
		err = proto.Unmarshal(message.Content, &eventList)
		if err != nil {
			return err
		}
		println("Received the following events: ----------")
		for _, event := range eventList.Events {
			fmt.Printf("Event: %v\n", *event)
			server.BroadcastToAll("event", *event)
		}
	}

	// Unsubscribe from events
	unSubscribeRequest := client_event_pb2.ClientEventsUnsubscribeRequest{}
	serializedRequest, err = proto.Marshal(&unSubscribeRequest)
	if err != nil {
		return err
	}
	corrId, err = zmqConnection.SendNewMsg(
		validator_pb2.Message_CLIENT_EVENTS_UNSUBSCRIBE_REQUEST,
		serializedRequest,
	)
	if err != nil {
		return err
	}
	// Wait for status
	_, unsubscribeResponse, err := zmqConnection.RecvMsgWithId(corrId)
	if err != nil {
		return err
	}
	eventUnsubscribeResponse := client_event_pb2.ClientEventsUnsubscribeResponse{}
	err = proto.Unmarshal(unsubscribeResponse.Content, &eventUnsubscribeResponse)
	if err != nil {
		return err
	}
	if eventUnsubscribeResponse.Status !=
		client_event_pb2.ClientEventsUnsubscribeResponse_OK {
		return errors.New("Client couldn't unsubscribe successfully")
	}
	return nil
}

func main() {

	// Entry point function for the client CLI.
	filters := []*events_pb2.EventFilter{&events_pb2.EventFilter{
		Key:         "address",
		MatchString: RAZORBACK_TP_ADDRESS_PREFIX + ".*",
		FilterType:  events_pb2.EventFilter_REGEX_ANY,
	}}

	// To listen to all events, there should not be any filters
	err := listenToEvents(filters)
	if err != nil {
		fmt.Printf("Error occurred %v\n", err)
	}
}
