package rpcservice

import (
	"fmt"
	"net/http"
	"github.com/goodsign/gosmsc"
)

//Service Definition
type SMSService struct {
	sender *gosmsc.Sender
}

func NewSMSService(sender *gosmsc.Sender) (*SMSService, error) {
	if sender == nil {
		return nil, fmt.Errorf("nil sender")
	}
	return &SMSService{sender}, nil
}

type Send_Args struct {
	Phone string
	Text  string
}
type Send_Reply struct {
	Message string
}

// SMSCInterface implementation
func (h *SMSService) Send(r *http.Request, msg *Send_Args, reply *Send_Reply) error {
	logger.Trace("")

	err := h.sender.Send(msg.Phone, msg.Text)
	if err != nil {
		reply.Message = err.Error()
		return err
	}
	reply.Message = "OK"
	return nil
}
