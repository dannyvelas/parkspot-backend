package email

type Sender interface {
	Send(body []byte) error
}
