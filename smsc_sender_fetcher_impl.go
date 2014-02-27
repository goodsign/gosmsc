package gosmsc

import (
	. "github.com/goodsign/gosmsc/contract"
)

// SenderFetcherImpl is a plain implementation of Sender and StatusFetcher interfaces that
// interacts with SMSC web gateway.
// Unlike SenderCheckerImpl, SenderFetcherImpl doesn't track or store anything, it is just
// a gateway caller without any side-effects.
type SenderFetcherImpl struct {
	sender        Sender
	statusFetcher StatusFetcher
}

func newSenderFetcherImplInternal(sender Sender, statusFetcher StatusFetcher) (*SenderFetcherImpl, error) {
	if sender == nil {
		return nil, logger.Error("sender cannot be nil")
	}

	if statusFetcher == nil {
		return nil, logger.Error("statusFetcher cannot be nil")
	}

	impl := new(SenderFetcherImpl)

	impl.sender = sender
	impl.statusFetcher = statusFetcher

	return impl, nil
}

func NewSenderFetcherImpl(opts *SmscClientOptions) (*SenderFetcherImpl, error) {
	sint, err := newSmsClientInternal(opts)
	if err != nil {
		return nil, err
	}
	return newSenderFetcherImplInternal(sint, sint)
}

func (c *SenderFetcherImpl) Send(phone string, text string) (*SendSMSResponse, error) {
	return c.sender.Send(phone, text)
}

func (c *SenderFetcherImpl) FetchStatus(id int64, phone string) (*CheckStatusResponse, error) {
	return c.statusFetcher.FetchStatus(id, phone)
}
