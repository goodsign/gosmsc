package rpcservice

import (
	"fmt"
	"github.com/goodsign/gosmsc"
	. "github.com/goodsign/gosmsc/contract"
	"net/http"
)

//Service Definition
type SMSService struct {
	senderChecker *gosmsc.SenderCheckerImpl
}

func NewSMSService(senderChecker *gosmsc.SenderCheckerImpl) (*SMSService, error) {
	if senderChecker == nil {
		return nil, fmt.Errorf("nil senderChecker")
	}
	return &SMSService{senderChecker}, nil
}

type Send_Args struct {
	Phone string
	Text  string
	Track bool
}
type Send_Reply struct {
	Id int64
}

// SMSCClientInterface implementation
func (h *SMSService) Send(r *http.Request, msg *Send_Args, reply *Send_Reply) error {
	logger.Trace("")

	id, err := h.senderChecker.Send(msg.Phone, msg.Text, msg.Track)
	if err != nil {
		return err
	}
	reply.Id = id
	return nil
}

type GetActualStatus_Args struct {
	Id int64
}
type GetActualStatus_Reply struct {
	Status *MessageStatus
}

// SMSCClientInterface implementation
func (h *SMSService) GetActualStatus(r *http.Request, msg *GetActualStatus_Args, reply *GetActualStatus_Reply) error {
	logger.Trace("")

	status, err := h.senderChecker.GetActualStatus(msg.Id)
	if err != nil {
		return err
	}
	reply.Status = status
	return nil
}
