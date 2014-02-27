package contract

// SenderChecker is an interface representing the ability to perform two main functions: sending sms and tracking them.
// Tracking here means the ability to poll SMSC gateway periodically.
//
// Both gosmsc.HttpSenderChecker and gosmsc/rpcservice/client.Client implement it, so it is easy to
// use this interface in your app and switch between local service implementation and a remote one.
type SenderChecker interface {
	// Send sends SMS via SMSC. Returns returned message id or an error.
	// If track flag is set, message must be added to the storage and tracker goroutine should start polling the
	// SMSC gateway and update its status until it is delivered. If track flag is not set, message is not tracked.
	Send(phone string, text string, track bool) (int64, error)

	// GetActualStatus gets the most actual (at current moment of time) message status.
	GetActualStatus(id int64) (*MessageStatus, error)
}

// Sender is an interface representing the ability to send sms using the SMSC gateway.
type Sender interface {
	Send(phone string, text string) (*SendSMSResponse, error) // Sends SMS via SMSC. Returns service response.
}

// StatusFetcher is an interface representing the ability to fetch sms status using the SMSC gateway.
type StatusFetcher interface {
	FetchStatus(id int64, phone string) (*CheckStatusResponse, error) // Gets current SMS status via SMSC. Returns service response.
}

// StatusContainer defines contract for tracked sms storage container.
type StatusContainer interface {
	Put(msgStatus *MessageStatus) error      // If status is already present, overwrite it. Overwise adds it.
	Get(msgId int64) (*MessageStatus, error) // Get status by its id, if present. If not, returns error.
	GetPending() ([]MessageStatus, error)    // Returns those with status not equal to MessageStatusComplete
}
