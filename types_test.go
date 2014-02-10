package gosmsc

import (
	"fmt"
	. "github.com/goodsign/gosmsc/contract"
	"sync"
	"time"
)

var mid sync.Mutex
var lid int64

func getNextMessageId() int64 {
	mid.Lock()
	defer mid.Unlock()
	lid++
	return lid
}

type smscTestClientOptions struct {
	invalidCreds       bool
	ioFaultRequested   bool
	expectedStatusCode MessageStatusCode
}

// smsTestClientInternal contains protocol-independent logic to connect to smsc service or its mock (used in tests).
type smsTestClientInternal struct {
	opts *smscTestClientOptions
}

func (c *smsTestClientInternal) Send(phone string, text string) (*SendSMSResponse, error) {
	if c.opts.ioFaultRequested {
		return nil, fmt.Errorf("Some io error")
	}
	if c.opts.invalidCreds {
		return &SendSMSResponse{"Invalid credentials", -123, 0}, nil
	}

	return &SendSMSResponse{"", 0, getNextMessageId()}, nil
}

func (c *smsTestClientInternal) FetchStatus(id int64, phone string) (*CheckStatusResponse, error) {
	if c.opts.ioFaultRequested {
		return nil, fmt.Errorf("Some io error")
	}
	if c.opts.invalidCreds {
		return &CheckStatusResponse{0, "", "", "", 0, "Invalid credentials", -123}, nil
	}

	return &CheckStatusResponse{int32(c.opts.expectedStatusCode), "02.01.2006 15:04:05", "", "", 0, "", 0}, nil
}

// messageStatusTestStorage is a default mgo implementation of the MessageStatusStorageInterface.
type messageStatusTestStorage struct {
	m    sync.Mutex
	msgs []MessageStatus
}

func newMessageStatusTestStorage() *messageStatusTestStorage {
	return &messageStatusTestStorage{}
}

func (ms *messageStatusTestStorage) Get(messageId int64) (*MessageStatus, error) {
	ms.m.Lock()
	defer ms.m.Unlock()

	for _, v := range ms.msgs {
		if v.MessageId == messageId {
			return &v, nil
		}
	}
	return nil, MessageNotFound
}

func (ms *messageStatusTestStorage) Put(message *MessageStatus) error {
	ms.m.Lock()
	defer ms.m.Unlock()

	if message == nil {
		return fmt.Errorf("message is nil")
	}
	for i, v := range ms.msgs {
		if v.MessageId == message.MessageId {
			ms.msgs[i] = *message
			return nil
		}
	}
	ms.msgs = append(ms.msgs, *message)
	return nil
}

func (ms *messageStatusTestStorage) GetPending() ([]MessageStatus, error) {
	ms.m.Lock()
	defer ms.m.Unlock()

	var p []MessageStatus
	for _, v := range ms.msgs {
		if v.StatusCode != MessageStatusComplete {
			p = append(p, v)
		}
	}
	return p, nil
}

func newTestSenderCheckerImpl(opts *smscTestClientOptions, updateInterval time.Duration) (*SenderCheckerImpl, error) {
	sint := &smsTestClientInternal{opts}
	return newSenderCheckerImplInternal(sint, sint, newMessageStatusTestStorage(), updateInterval)
}
