package eventsource

import (
	"container/list"
	"net/http"
	"sync"
)

type Stream struct {
	clients           list.List
	listLock          sync.RWMutex
	broadcast         chan *Event
	shutdownWait      sync.WaitGroup
	clientConnectHook func(*http.Request, *Client)
}

type registeredClient struct {
	c      *Client
	topics map[string]bool
}

func New() *Stream {
	s := &Stream{}
	go s.run()
	return s
}

// Register adds a client to the stream to receive all broadcast
// messages. Has no effect if the client is already registered.
func (s *Stream) Register(c *Client) {

	// see if the client has been registered
	if cli := s.getClient(c); cli != nil {
		return
	}

	// append new client
	s.addClient(c)
}

// Broadcast sends the event to all clients registered on this stream.
func (s *Stream) Broadcast(e *Event) {

	for element := s.clients.Front(); element != nil; element.Next() {
		cli := element.Value.(*registeredClient)
		cli.c.Send(e)
	}
}

// Subscribe add the client to the list of clients receiving publications
// to this topic. Subscribe will also Register an unregistered
// client.
func (s *Stream) Subscribe(topic string, c *Client) {

	// see if the client is registered
	cli := s.getClient(c)

	// register if not
	if cli == nil {
		cli = s.addClient(c)
	}

	cli.topics[topic] = true
}

// Unsubscribe removes clients from the topic, but not from broadcasts.
func (s *Stream) Unsubscribe(topic string, c *Client) {
	cli := s.getClient(c)
	if cli == nil {
		return
	}
	cli.topics[topic] = false
}

// Publish sends the event to clients that have subscribed to the given topic.
func (s *Stream) Publish(topic string, e *Event) {
	for element := s.clients.Front(); element != nil; element.Next() {
		cli := element.Value.(*registeredClient)
		if cli.topics[topic] {
			cli.c.Send(e)
		}
	}
}

func (s *Stream) Shutdown() {

}

func (s *Stream) CloseTopic(topic string) {

}

func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func (s *Stream) TopicHandler(topic string) http.HandlerFunc {

}

// Register a function to be called when a client connects to this stream's
// HTTP handler
func (s *Stream) ClientConnectHook(fn func(*http.Request, *Client)) {
	s.clientConnectHook = fn
}

func (s *Stream) run() {

	for {
		select {

		case ev, ok := <-s.broadcast:

			// end of the broadcast channel indicates a shutdown
			if !ok {
				//s.closeAll()
				s.shutdownWait.Done()
				return
			}

			// otherwise normal message
			s.sendAll(ev)

		}
	}
}

func (s *Stream) sendAll(ev *Event) {

}

func (s *Stream) getClient(c *Client) *registeredClient {
	if s.clients.Len() > 0 {
		// ensure client is not already registered
		s.listLock.RLock()

		listItem := s.clients.Front()
		if regCli := listItem.Value.(*registeredClient); regCli.c == c {
			// client found
			s.listLock.RUnlock()
			return regCli
		}
	}

	return nil
}

func (s *Stream) addClient(c *Client) *registeredClient {
	s.listLock.Lock()

	s.clients.PushBack(&registeredClient{
		c:      c,
		topics: make(map[string]bool),
	})

	s.listLock.Unlock()
}