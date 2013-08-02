package gosmsc

import (
	"fmt"
	"net/http"
)

// SenderOptions encapsulates configuration used to send sms messages using smsc.ru
type SenderOptions struct {
	User         string `json:"user"`
	Password     string `json:"pwd"`
}

// Sender is used to perform different smsc.ru calls using the same sender configuration (user, pwd, etc.)
type Sender struct {
	opts *SenderOptions
}

func NewSender(opts *SenderOptions) (*Sender, error) {
	if len(opts.User) == 0 {
		return nil, fmt.Errorf("Nil length user")
	}
	if len(opts.Password) == 0 {
		return nil, fmt.Errorf("Nil length password")
	}
	return &Sender{opts}, nil
}

// SendSMS uses 'send' API to send SMS with specified phone, text, and configuration
func (sender *Sender) Send(phone string, text string) error {
	resp, err := http.Get(fmt.Sprintf("https://smsc.ru/sys/send.php?login=%s&psw=%s&charset=utf-8&phones=%s&mes=%s",
									  sender.opts.User, sender.opts.Password, phone, text))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}