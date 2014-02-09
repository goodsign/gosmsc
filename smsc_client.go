package gosmsc

import (
	. "github.com/goodsign/gosmsc/contract"
)

// SmscClientOptions encapsulates configuration used to send sms messages using smsc.ru
type SmscClientOptions struct {
	User     string `json:"user"`
	Password string `json:"pwd"`
}

// SmscClient is used to perform different smsc.ru calls using the same SmscClient configuration (user, pwd, etc.)
type SmscClient struct {
	smsClientInternal
}

func NewSmscClient(opts *SmscClientOptions) (*SmscClient, error) {
	swrap, err := newDefaultSmscWrapper(opts)
	if err != nil {
		return nil, err
	}
	return &SmscClient{smsClientInternal{swrap}}, nil
}

// Send uses HTTPS API to send SMS with specified phone, text, and configuration.
// Returns service response bytes.
func (c *SmscClient) Send(phone string, text string) (*MessageStatus, error) {
	return c.send(phone, text)
}

// GetStatus uses HTTPS API to get sent SMS status for the specified message id
// and phone.
// Returns service response bytes.
func (c *SmscClient) GetStatus(id int64, phone string) ([]byte, error) {
	return c.callGetStatusService(id, phone)
}
