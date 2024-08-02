package ws

type Message struct {
	Op   int
	Uid  string
	Data interface{}
}
