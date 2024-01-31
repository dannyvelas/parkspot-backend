package email

var _ Sender = (*MockService)(nil)

type MockService struct {
	expectedErr error
}

func NewMockService(err error) MockService {
	return MockService{err}
}

func (s MockService) Send(body []byte) error {
	return s.expectedErr
}
