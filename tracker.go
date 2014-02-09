package gosmsc

import (
	"encoding/json"
	"fmt"
	. "github.com/goodsign/gosmsc/contract"
	"sync"
	"time"
)

const (
	DefaultUpdateInterval = time.Minute
)

// MessageTracker represents a running goroutine that polls smsc service to track status of sent messages
// which status is pending. This object is created when a goroutine is started by StartTracking and
// can be used to stop the goroutine using Close func. After goroutine is stopped by Close() this
// object cannot be used anymore.
type MessageTracker struct {
	stopM       sync.Mutex
	storage     MessageStatusStorageInterface
	smsc        SMSCInterface
	stopped     bool
	stopChannel chan bool // Used to signal the polling goroutine to stop and finish
}

// StartTracking creates a new tracker for the specified storage and starts the tracking process
// in a separate goroutine. Uses 'smsc' argument to call the SMSC service when updates for pending
// messages are needed.
// To stop it, call Close on the returned tracker instance.
func StartTracking(storage MessageStatusStorageInterface, smsc SMSCInterface, updateInterval time.Duration) (tracker *MessageTracker, e error) {
	if storage == nil {
		return nil, fmt.Errorf("Message tracker storage cannot be nil")
	}
	if smsc == nil {
		return nil, fmt.Errorf("Message tracker smsc parameter cannot be nil")
	}
	tracker = &MessageTracker{sync.Mutex{}, storage, smsc, false, make(chan bool, 1)}

	go func(t *MessageTracker) {
		ticker := time.NewTicker(updateInterval)
		for !tracker.stopped {
			select {
			case <-ticker.C:
				t.checkPending()
			case <-t.stopChannel:
			}
		}
	}(tracker)

	return
}

// IsStopped returns true if the tracker goroutine was stopped by the Stop func and the tracker is unusable anymore.
func (t MessageTracker) IsStopped() bool {
	t.stopM.Lock()
	defer t.stopM.Unlock()
	return t.stopped
}

// Stop stops the polling goroutine and closes the tracker object. A stopped tracker
// cannot be used anymore.
//
// NOTE 1: Goroutine doesn't terminate immediately (it can be processing pending issues), but
// Stop func is non blocking itself.
func (t MessageTracker) Stop() error {
	t.stopM.Lock()
	defer t.stopM.Unlock()
	if t.stopped {
		return fmt.Errorf("Already stopped")
	}
	t.stopped = true
	t.stopChannel <- true
	return nil
}

func (t MessageTracker) checkPending() error {
	pendingMessages, err := t.storage.GetPending()
	if err != nil {
		return logger.Error(err)
	}

	for _, message := range pendingMessages {
		logger.Debug("Checking message %v", message.MessageId)
		responseBytes, err := t.smsc.GetStatus(message.MessageId, message.Phone)
		if err != nil {
			logger.Error(err)
			continue
		}

		output := new(checkStatusResponse)
		err = json.Unmarshal(responseBytes, &output)
		if err != nil {
			return logger.Error(err)
		}

		if output.Error != "" {
			return logger.Error(fmt.Errorf("[%v] %s", output.ErrorCode, output.Error))
		}

		message.StatusCode = MessageStatusCode(output.StatusCode)
		message.Operator = output.Operator
		message.Region = output.Region
		message.StatusErrorCode = output.StatusErrorCode

		statusUpdatedAt, err := time.Parse("02.01.2006 15:04:05", output.StatusDate)
		if err != nil {
			logger.Error(err)
		} else {
			message.StatusUpdatedAt = statusUpdatedAt.Local()
		}

		err = t.storage.Put(&message)
		if err != nil {
			logger.Error(err)
			continue
		}
	}
	return nil
}

// Used to unmarshal sms status response.
type checkStatusResponse struct {
	StatusCode      int32  `json:"status"`
	StatusDate      string `json:"last_date"`
	Operator        string `json:"operator"`
	Region          string `json:"region"`
	StatusErrorCode int32  `json:"err"`
	Error           string `json:"error"`
	ErrorCode       int32  `json:"error_code"`
}
