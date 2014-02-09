package gosmsc

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/goodsign/gosmsc/contract"
	"io/ioutil"
	"net/http"
)

// smscWrapper is used to provide different protocols to communicate with
// SMSC or mock it for tests.
type smscWrapper interface {
	sendSms(phone string, text string) ([]byte, error)
	callGetStatusService(id int64, phone string) ([]byte, error)
}

// defaultSmscWrapper represents the default http protocol to communicate with smsc service.
type defaultSmscWrapper struct {
	opts *SmscClientOptions
}

func newDefaultSmscWrapper(opts *SmscClientOptions) (*defaultSmscWrapper, error) {
	if len(opts.User) == 0 {
		return nil, fmt.Errorf("Nil length user")
	}
	if len(opts.Password) == 0 {
		return nil, fmt.Errorf("Nil length password")
	}
	return &defaultSmscWrapper{opts}, nil
}

func (s defaultSmscWrapper) get(path string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("https://smsc.ru/%s", path))
	if err != nil {
		return nil, logger.Error(err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, logger.Error(err)
	}
	return respBytes, nil
}

func (w defaultSmscWrapper) sendSms(phone string, text string) ([]byte, error) {
	return w.get(fmt.Sprintf("sys/send.php?login=%s&psw=%s&charset=utf-8&phones=%s&mes=%s",
		w.opts.User, w.opts.Password, phone, text))
}

func (w defaultSmscWrapper) callGetStatusService(id int64, phone string) ([]byte, error) {
	return w.get(fmt.Sprintf("sys/status.php?login=%s&psw=%s&phone=%s&id=%v&fmt=3&all=2&charset=utf-8",
		w.opts.User, w.opts.Password, phone, id))
}

// smsClientInternal contains protocol-independent logic to connect to smsc service or its mock (used in tests).
type smsClientInternal struct {
	wrapper smscWrapper
}

func (c *smsClientInternal) send(phone string, text string) (*MessageStatus, error) {
	respBytes, err := c.wrapper.sendSms(phone, text)
	if err != nil {
		return nil, logger.Error(err)
	}

	output := new(sendSMSResponse)
	err = json.Unmarshal(respBytes, &output)
	if err != nil {
		return nil, logger.Error(err)
	}

	if output.Error != "" {
		err = errors.New(fmt.Sprintf("[%v] %s", output.ErrorCode, output.Error))
		return nil, logger.Error(err)
	}

	return NewUnknownMessageStatus(output.Id, phone), nil
}

func (c *smsClientInternal) callGetStatusService(id int64, phone string) ([]byte, error) {
	return c.wrapper.callGetStatusService(id, phone)
}

// sendSMSResponse is used to unmarshal the response from server on the 'send sms' action.
// See SmscClient.Send.
type sendSMSResponse struct {
	Error     string `json:"error"`
	ErrorCode int32  `json:"error_code"`
	Id        int64  `json:"id"`
}
