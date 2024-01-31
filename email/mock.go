package email

var _ Sender = MockService{}

type MockService struct {
}

func NewMockService() MockService {
	return MockService{}
}

func (s MockService) Send(body []byte) error {
	return nil
}
