package event

type Listener interface {
	Listen(output chan<- []byte) error
	GetListenerId() string
	Stop() error
}
