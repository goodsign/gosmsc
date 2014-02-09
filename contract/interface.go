package contract

// SMSCInterface defines smsc service contract.
// Both gosmsc.Sender and gosmsc/rpcservice/client.Client implement it, so it is easy to
// use this interface in your app and switch between local sender and a remote sender service.
type SMSCInterface interface {
	Send(phone string, text string) (*MessageStatus, error) // Sends SMS via SMSC. Returns service response bytes.
	GetStatus(id int64, phone string) ([]byte, error)       // Gets SMS status via SMSC. Returns service response bytes.
}

// MessageStatusStorageInterface defines contract for tracked sms storage container.
type MessageStatusStorageInterface interface {
	Put(msgStatus *MessageStatus) error      // If status is already present, overwrite it. Overwise adds it.
	Get(msgId int64) (*MessageStatus, error) // Get status by its id, if present. If not, returns error.
	GetPending() ([]MessageStatus, error)    // Returns those with status not equal to MessageStatusComplete
	//GetAll() ([]MessageStatus, error)      // Returns all message statuses.
}
