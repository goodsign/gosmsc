package gosmsc

import (
	"errors"
	"fmt"
	. "github.com/goodsign/gosmsc/contract"
	mgohelper "github.com/goodsign/goutils/mgo"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var (
	messagesCollection = "messages"
	MessageNotFound    = errors.New("Message not found")
	MessageSuspended   = errors.New("Message suspended")
)

// MessageStatusMgoStorage is a default mgo implementation of the MessageStatusStorageInterface.
type MessageStatusMgoStorage struct {
	h *mgohelper.DbHelper
}

func NewMessageStatusMgoStorage(dbHelper *mgohelper.DbHelper) (*MessageStatusMgoStorage, error) {
	if dbHelper == nil {
		return nil, fmt.Errorf("dbHelper is nil")
	}
	return &MessageStatusMgoStorage{dbHelper}, nil
}

func (ms *MessageStatusMgoStorage) Get(messageId int64) (*MessageStatus, error) {
	logger.Tracef("messageId: '%d'", messageId)
	c, s := ms.h.C(messagesCollection)
	defer s.Close()

	message := new(MessageStatus)
	err := c.Find(bson.M{"messageid": messageId}).One(message)
	if err != nil {
		if err != mgo.ErrNotFound {
			return nil, logger.Error(err)
		}
		return nil, MessageNotFound
	}

	return message, nil
}

func (ms *MessageStatusMgoStorage) Put(message *MessageStatus) error {
	if message == nil {
		return logger.Errorf("message is nil")
	}
	logger.Tracef("message id: '%d'", message.MessageId)

	c, s := ms.h.C(messagesCollection)
	defer s.Close()
	_, err := c.Upsert(bson.M{"messageid": message.MessageId, "phone": message.Phone}, message)
	return err
}

func (ms *MessageStatusMgoStorage) GetPending() ([]MessageStatus, error) {
	logger.Trace("")

	c, s := ms.h.C(messagesCollection)
	defer s.Close()

	var messages []MessageStatus
	err := c.Find(bson.M{"statuscode": bson.M{"$ne": MessageStatusComplete}, "statuserrorcode": 0}).All(&messages)
	if err != nil {
		return nil, logger.Error(err)
	}
	return messages, nil
}
