package gosmsc

import (
	"encoding/json"
	"fmt"
	. "github.com/goodsign/gosmsc/contract"
	"io/ioutil"
	"net/http"
)

// smsClientInternal contains protocol-independent logic to connect to smsc service or its mock (used in tests).
type smsClientInternal struct {
	opts *SmscClientOptions
}

func newSmsClientInternal(opts *SmscClientOptions) (*smsClientInternal, error) {
	if len(opts.User) == 0 {
		return nil, fmt.Errorf("Nil length user")
	}
	if len(opts.Password) == 0 {
		return nil, fmt.Errorf("Nil length password")
	}
	return &smsClientInternal{opts}, nil
}

func (c *smsClientInternal) get(path string) ([]byte, error) {
	getPath := fmt.Sprintf("https://smsc.ru/%s", path)
	logger.Infof("GET: '%s'", getPath)
	resp, err := http.Get(getPath)
	if err != nil {
		return nil, logger.Error(err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	logger.Debugf("Server response:\n %s", string(respBytes))
	if err != nil {
		return nil, logger.Error(err)
	}
	return respBytes, nil
}

func (c *smsClientInternal) Send(phone string, text string) (*SendSMSResponse, error) {
	respBytes, err := c.get(fmt.Sprintf("sys/send.php?login=%s&psw=%s&charset=utf-8&phones=%s&mes=%s&fmt=3",
		c.opts.User, c.opts.Password, phone, text))
	if err != nil {
		return nil, logger.Error(err)
	}

	output := new(SendSMSResponse)
	err = json.Unmarshal(respBytes, &output)
	if err != nil {
		return nil, logger.Error(err)
	}
	return output, nil
}

func (c *smsClientInternal) FetchStatus(id int64, phone string) (*CheckStatusResponse, error) {
	respBytes, err := c.get(fmt.Sprintf("sys/status.php?login=%s&psw=%s&phone=%s&id=%v&fmt=3&all=2&charset=utf-8",
		c.opts.User, c.opts.Password, phone, id))
	if err != nil {
		return nil, logger.Error(err)
	}
	output := new(CheckStatusResponse)
	err = json.Unmarshal(respBytes, &output)
	if err != nil {
		return nil, logger.Error(err)
	}
	return output, nil
}
