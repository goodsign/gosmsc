package contract

import (
	"time"
)

// Status code of a tracked sms message.
type MessageStatusCode int32

const (
	MessageStatusCodeUnknown = -999
	MessageStatusJustSent    = -2
	MessageStatusComplete    = 1
)

// MessageStatus represents status of a message that was sent using the smsc service
// and assigned an ID. MessageStatus can be stored in the database and updated when
// new information about its status is retrieved from the smsc service.
type MessageStatus struct {
	MessageId       int64 // Assigned when message is created. See Sender.Send
	Phone           string
	CreatedAt       time.Time
	StatusUpdatedAt time.Time // Time when this struct was modified last time
	StatusCode      MessageStatusCode
	Operator        string
	Region          string
	StatusErrorCode int32 // Not null if server returned an error code during the last update
}

// NewUnknownMessageStatus creates a new message which status is unknown. E.g. just created message.
// Unknown status represents status information about the message that was just sent via the sms service,
// but which code was not retrieved yet.
func NewUnknownMessageStatus(messageId int64, phone string) *MessageStatus {
	return &MessageStatus{messageId, phone, time.Now(), time.Now(), MessageStatusCodeUnknown, "", "", 0}
}

// CheckStatusResponse is used to unmarshal server response for status checking request.
type CheckStatusResponse struct {
	StatusCode      int32  `json:"status"`
	StatusDate      string `json:"last_date"`
	Operator        string `json:"operator"`
	Region          string `json:"region"`
	StatusErrorCode int32  `json:"err"`
	Error           string `json:"error"`
	ErrorCode       int32  `json:"error_code"`
}

// SendSMSResponse is used to unmarshal the response from server on the 'send sms' action.
// See SmscClient.Send.
type SendSMSResponse struct {
	Error     string `json:"error"`
	ErrorCode int32  `json:"error_code"`
	Id        int64  `json:"id"`
}
