package client

import (
	. "github.com/goodsign/gosmsc/contract"
	service "github.com/goodsign/gosmsc/rpcservice"
	"github.com/goodsign/goutils/jsonrpc"
	"time"
)

const SmscRpcServiceName = "SMSService."

// EmptyStruct is used in funcs where logicaly no input parameters or return values (or both) are needed, but
// the signature contains them as obligatory func arguments.
type EmptyStruct struct{}

// DbServiceClient provides convenient interface to JSON-RPC client for DB. It hides all
// transport level context and exposes only call logic related inputs and outputs.
type SmscRpcServiceClient struct {
	*jsonrpc.ServiceClient
}

// NewSmscRpcServiceClient creates a new SmscRpcServiceClient working with
// the service at the specified address with specified params:
//   * Address must be a full network address with port, e.g. 'http://localhost:8080/service'.
//   * Retry count/timeout specify retry strategy when calling any server method. Each call
//     can finally fail (return transport error) only after 'retryCount' fails with 'retryTimeout'
//     interval between them.
func NewSmscRpcServiceClient(address string, retryCount int, retryTimeout time.Duration) (*SmscRpcServiceClient, error) {
	db := new(SmscRpcServiceClient)
	c, err := jsonrpc.NewServiceClient(address, retryCount, retryTimeout)

	if err != nil {
		return nil, err
	}

	db.ServiceClient = c
	return db, nil
}

//------------------------------------------------
// ▢ Send
//------------------------------------------------

func (client *SmscRpcServiceClient) Send(phone string, text string, track bool) (int64, error) {
	args := service.Send_Args{phone, text, track}
	var r service.Send_Reply

	e := client.GetResult(SmscRpcServiceName+"Send", &args, &r)
	if e != nil {
		return 0, e
	}
	return r.Id, nil
}

//------------------------------------------------
// ▢ GetActualStatus
//------------------------------------------------

func (client *SmscRpcServiceClient) GetActualStatus(id int64) (*MessageStatus, error) {
	args := service.GetActualStatus_Args{id}
	var r service.GetActualStatus_Reply

	e := client.GetResult(SmscRpcServiceName+"GetActualStatus", &args, &r)
	if e != nil {
		return nil, e
	}
	return r.Status, nil
}
