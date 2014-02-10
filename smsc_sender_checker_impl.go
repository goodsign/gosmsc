package gosmsc

import (
	. "github.com/goodsign/gosmsc/contract"
	"time"
)

// SmscClientOptions encapsulates configuration used to send sms messages using smsc.ru
type SmscClientOptions struct {
	User     string `json:"user"`
	Password string `json:"pwd"`
}

// HttpSenderChecker provides the functionality to send sms and track its status.
//
// It consists of a sender, status getter, message storage, and a tracker goroutine.
// When sms is sent using 'Send', SenderChecker implementation adds a 'pending' entry to the storage. Then the
// tracker goroutine starts polling the smsc gateway using the status getter. It polls it until message is delivered
// and updates the message status in the storage each time.
//
// It can be also used without tracking functionality. See 'track' flag in the Send func.
type SenderCheckerImpl struct {
	sender        Sender
	storage       StatusContainer
	statusFetcher StatusFetcher
	tracker       *MessageTracker
}

func newSenderCheckerImplInternal(sender Sender, statusFetcher StatusFetcher, storage StatusContainer, updateInterval time.Duration) (*SenderCheckerImpl, error) {
	if sender == nil {
		return nil, logger.Error("sender cannot be nil")
	}

	if statusFetcher == nil {
		return nil, logger.Error("statusFetcher cannot be nil")
	}

	if storage == nil {
		return nil, logger.Error("storage cannot be nil")
	}

	if updateInterval <= 0 {
		return nil, logger.Error("updateInterval cannot be zero or negative")
	}

	impl := new(SenderCheckerImpl)
	impl.sender = sender
	impl.storage = storage
	impl.statusFetcher = statusFetcher

	t, err := StartTracking(storage, statusFetcher, updateInterval)
	if err != nil {
		return nil, err
	}

	impl.tracker = t

	return impl, nil
}

func NewSenderCheckerImpl(opts *SmscClientOptions, storage StatusContainer, updateInterval time.Duration) (*SenderCheckerImpl, error) {
	sint, err := newSmsClientInternal(opts)
	if err != nil {
		return nil, err
	}
	return newSenderCheckerImplInternal(sint, sint, storage, updateInterval)
}

func (c *SenderCheckerImpl) Send(phone string, text string, track bool) (int64, error) {
	output, err := c.sender.Send(phone, text)
	if err != nil {
		return -1, logger.Error(err)
	}
	if output.Error != "" {
		return -1, logger.Errorf("[%v] %s", output.ErrorCode, output.Error)
	}

	if track {
		st := NewUnknownMessageStatus(output.Id, phone)
		err = c.storage.Put(st)
		if err != nil {
			return -1, logger.Error(err)
		}
	}
	return output.Id, nil
}

func (c *SenderCheckerImpl) GetActualStatus(id int64) (*MessageStatus, error) {
	return c.storage.Get(id)
}
