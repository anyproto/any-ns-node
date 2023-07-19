package anynsrpc

import (
	"context"
	"errors"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/anyproto/any-sync/util/crypto"
	"github.com/anyproto/anyns-node/config"

	as "github.com/anyproto/anyns-node/pb/anyns_api_server"
)

const CName = "anyns.rpc"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsRpc{}
}

// consensusRpc implements consensus rpc server
type anynsRpc struct {
	contractsConfig config.Contracts
}

func (arpc *anynsRpc) Init(a *app.App) (err error) {
	arpc.contractsConfig = a.MustComponent(config.CName).(*config.Config).GetContracts()

	return as.DRPCRegisterAnyns(a.MustComponent(server.CName).(server.DRPCServer), arpc)
}

func (arpc *anynsRpc) Name() (name string) {
	return CName
}

func (arpc *anynsRpc) GetOperationStatus(ctx context.Context, in *as.GetOperationStatusRequest) (*as.OperationResponse, error) {
	// TODO: get status from the queue
	// for now, just return completed
	var resp as.OperationResponse
	resp.OperationId = in.OperationId
	resp.OperationState = as.OperationState_Completed

	return &resp, nil
}

func (arpc *anynsRpc) IsNameAvailable(ctx context.Context, in *as.NameAvailableRequest) (*as.NameAvailableResponse, error) {
	return IsNameAvailable(ctx, in, &arpc.contractsConfig)
}

func (arpc *anynsRpc) NameRegister(ctx context.Context, in *as.NameRegisterRequest) (*as.OperationResponse, error) {
	var resp as.OperationResponse // TODO: make non-blocking and save to queue
	resp.OperationId = 1          // TODO: increase the operation ID

	err := NameRegister(ctx, in, &arpc.contractsConfig)

	if err != nil {
		log.Fatal("can not register name", zap.Error(err))
		resp.OperationState = as.OperationState_Error
		return &resp, err
	}

	resp.OperationState = as.OperationState_Completed
	return &resp, err
}

func VerifyIdentity(in *as.NameRegisterSignedRequest, ownerAnyAddress string) error {
	// convert ownerAnyAddress to array of bytes
	arr := []byte(ownerAnyAddress)

	ownerAnyIdentity, err := crypto.UnmarshalEd25519PublicKeyProto(arr)
	if err != nil {
		return err
	}

	res, err := ownerAnyIdentity.Verify(in.Payload, in.Signature)
	if err != nil || !res {
		return errors.New("signature is different")
	}

	// identity is OK
	return nil
}

func (arpc *anynsRpc) NameRegisterSigned(ctx context.Context, in *as.NameRegisterSignedRequest) (*as.OperationResponse, error) {
	var resp as.OperationResponse // TODO: make non-blocking and save to queue
	resp.OperationId = 1          // TODO: increase the operation ID

	// 1 - unmarshal the signed request
	var nrr as.NameRegisterRequest
	err := proto.Unmarshal(in.Payload, &nrr)
	if err != nil {
		resp.OperationState = as.OperationState_Error
		log.Fatal("can not unmarshal NameRegisterRequest", zap.Error(err))
		return &resp, err
	}

	// 2 - check signature
	err = VerifyIdentity(in, nrr.OwnerAnyAddress)
	if err != nil {
		resp.OperationState = as.OperationState_Error
		log.Fatal("identity is different", zap.Error(err))
		return &resp, err
	}

	// 3 - finally call function
	err = NameRegister(ctx, &nrr, &arpc.contractsConfig)

	if err != nil {
		log.Fatal("can not register name", zap.Error(err))
		resp.OperationState = as.OperationState_Error
		return &resp, err
	}

	resp.OperationState = as.OperationState_Completed
	return &resp, err
}

func (arpc *anynsRpc) NameUpdate(ctx context.Context, in *as.NameUpdateRequest) (*as.OperationResponse, error) {
	// TODO:
	return nil, nil
}
