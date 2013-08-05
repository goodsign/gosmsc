package contract

// SMSCInterface defines smsc service contract.
// Both gosmsc.Sender and gosmsc/rpcservice/client.Client implement it, so it is easy to
// use this interface in your app and switch between local sender and a remote sender service.
type SMSCInterface interface {
	Send(phone string, text string) error
}