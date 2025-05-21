package dblisteners

type DBListeners interface {
	Start() error
	SetInput(ch <-chan interface{})
	SetOutput(ch chan<- interface{})
}
