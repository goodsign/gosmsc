package gosmsc

import (
	"fmt"
	. "github.com/goodsign/gosmsc/contract"
	"sync"
	"time"
)

const (
	DefaultUpdateInterval = time.Minute
)

// MessageTracker represents a running goroutine that polls SMSC service to track status of sent messages
// which status is pending. This object is created when a goroutine is started by StartTracking and
// can be used to stop the goroutine using Close func. After goroutine is stopped by Close() this
// object cannot be used anymore.
type MessageTracker struct {
	stopM         sync.Mutex
	storage       StatusContainer
	statusFetcher StatusFetcher
	tickerForTest chan bool // Used to create artificial ticks from tests
	stopped       bool
	stopChannel   chan bool // Used to signal the polling goroutine to stop and finish
}

// StartTracking creates a new tracker for the specified storage and starts the tracking process
// in a separate goroutine. Uses 'statusFetcher' argument to call the SMSC service when updates for pending
// messages are needed.
// To stop it, call Close on the returned tracker instance.
func StartTracking(storage StatusContainer, statusFetcher StatusFetcher, updateInterval time.Duration) (tracker *MessageTracker, e error) {
	if updateInterval <= 0 {
		return nil, fmt.Errorf("updateInterval cannot be zero or negative")
	}
	if storage == nil {
		return nil, fmt.Errorf("Message tracker storage cannot be nil")
	}
	if statusFetcher == nil {
		return nil, fmt.Errorf("Message tracker statusFetcher parameter cannot be nil")
	}
	tracker = &MessageTracker{sync.Mutex{}, storage, statusFetcher, make(chan bool), false, make(chan bool, 1)}

	go func(t *MessageTracker) {
		ticker := time.NewTicker(updateInterval)
		for !t.IsStopped() {
			select {
			case <-t.tickerForTest:
				t.checkPending()
			case <-ticker.C:
				t.checkPending()
			case <-t.stopChannel:
			}
		}
	}(tracker)

	return
}

// IsStopped returns true if the tracker goroutine was stopped by the Stop func and the tracker is unusable anymore.
func (t *MessageTracker) IsStopped() bool {
	t.stopM.Lock()
	defer t.stopM.Unlock()
	return t.stopped
}

// Stop stops the polling goroutine and closes the tracker object. A stopped tracker
// cannot be used anymore.
//
// NOTE 1: Goroutine doesn't terminate immediately (it can be processing pending issues), but
// Stop func is non blocking itself.
func (t *MessageTracker) Stop() error {
	t.stopM.Lock()
	defer t.stopM.Unlock()
	if t.stopped {
		return fmt.Errorf("Already stopped")
	}
	t.stopped = true
	t.stopChannel <- true
	close(t.tickerForTest)
	close(t.stopChannel)
	return nil
}

func (t *MessageTracker) checkPending() error {
	pendingMessages, err := t.storage.GetPending()
	if err != nil {
		return logger.Error(err)
	}

	for _, message := range pendingMessages {
		logger.Debug("Checking message %v", message.MessageId)
		output, err := t.statusFetcher.FetchStatus(message.MessageId, message.Phone)
		if err != nil {
			logger.Error(err)
			continue
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
