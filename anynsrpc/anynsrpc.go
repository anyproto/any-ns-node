package anynsrpc

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/anyproto/anyns-node/config"

	contracts "github.com/anyproto/anyns-node/contracts"
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
	contracts       contracts.Service
}

func (arpc *anynsRpc) Init(a *app.App) (err error) {
	arpc.contractsConfig = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.Service)

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
	// 0 - create connection
	conn, err := arpc.contracts.CreateEthConnection()
	if err != nil {
		log.Fatal("failed to connect to geth", zap.Error(err))
		return nil, err
	}

	// 1 - convert to name hash
	nh, err := contracts.NameHash(in.FullName)
	if err != nil {
		log.Fatal("can not convert FullName to namehash", zap.Error(err))
		return nil, err
	}

	// 2 - call contract's method
	log.Info("getting owner for name", zap.String("FullName", in.GetFullName()))
	addr, err := arpc.contracts.GetOwnerForNamehash(conn, nh)
	if (err != nil) || (addr == nil) {
		log.Fatal("failed to get owner", zap.Error(err))
		return nil, err
	}

	// 3 - covert to result
	// the owner can be NameWrapper
	log.Info("received owner address", zap.String("Owner addr", addr.Hex()))

	var res as.NameAvailableResponse
	var addrEmpty = common.Address{}

	if *addr == addrEmpty {
		log.Info("name is not registered yet...")
		res.Available = true
		return &res, nil
	}

	// 4 - if name is not available, then get additional info
	log.Info("name is NOT available...Getting additional info")
	ea, aa, si, err := arpc.contracts.GetAdditionalNameInfo(conn, *addr, in.GetFullName())
	if err != nil {
		log.Fatal("failed to get additional info", zap.Error(err))
		return nil, err
	}

	log.Info("name is already registered...")
	res.Available = false
	res.OwnerEthAddress = ea
	res.OwnerAnyAddress = aa
	res.SpaceId = si
	return &res, nil
}

func (arpc *anynsRpc) NameRegister(ctx context.Context, in *as.NameRegisterRequest) (*as.OperationResponse, error) {
	var resp as.OperationResponse // TODO: make non-blocking and save to queue
	resp.OperationId = 1          // TODO: increase the operation ID

	err := arpc.nameRegister(ctx, in)

	if err != nil {
		log.Fatal("can not register name", zap.Error(err))
		resp.OperationState = as.OperationState_Error
		return &resp, err
	}

	resp.OperationState = as.OperationState_Completed
	return &resp, err
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
	err = arpc.nameRegister(ctx, &nrr)

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

func (arpc *anynsRpc) nameRegister(ctx context.Context, in *as.NameRegisterRequest) error {
	var registrantAccount common.Address = common.HexToAddress(in.OwnerEthAddress)

	// TODO:
	// 0 - check all parameters

	conn, err := arpc.contracts.CreateEthConnection()
	if err != nil {
		log.Fatal("failed to connect to geth", zap.Error(err))
		return err
	}

	// 1 - connect to geth
	controller, err := arpc.contracts.ConnectToController(conn)
	if err != nil {
		log.Fatal("failed to connect to contract", zap.Error(err))
		return err
	}

	// 2 - get a name's first part
	// TODO: normalize string
	nameFirstPart := contracts.RemoveTLD(in.FullName)

	// 3 - calculate a commitment
	secret, err := contracts.GenerateRandomSecret()
	if err != nil {
		log.Fatal("can not generate random secret", zap.Error(err))
		return err
	}

	commitment, err := arpc.contracts.MakeCommitment(
		nameFirstPart,
		registrantAccount,
		secret,
		controller,
		in.GetFullName(),
		in.GetOwnerAnyAddress(),
		in.GetSpaceId())

	if err != nil {
		log.Fatal("can not calculate a commitment", zap.Error(err))
		return err
	}

	authOpts, err := arpc.contracts.GenerateAuthOptsForAdmin(conn)
	if err != nil {
		log.Fatal("can not get auth params for admin", zap.Error(err))
		return err
	}

	// 2 - send a commit transaction from Admin
	tx, err := arpc.contracts.Commit(
		authOpts,
		commitment,
		controller)
	if err != nil {
		log.Fatal("can not Commit tx", zap.Error(err))
		return err
	}

	// 3 - wait for tx to be mined
	arpc.contracts.WaitMined(ctx, conn, tx)

	txRes := arpc.contracts.CheckTransactionReceipt(conn, tx.Hash())
	if !txRes {
		log.Warn("commit TX failed", zap.Error(err))
		return errors.New("commit tx failed")
	}

	// update nonce again...
	authOpts, err = arpc.contracts.GenerateAuthOptsForAdmin(conn)
	if err != nil {
		log.Fatal("can not get auth params for admin", zap.Error(err))
		return err
	}

	// 4 - now send register tx
	tx, err = arpc.contracts.Register(
		authOpts,
		nameFirstPart,
		registrantAccount,
		secret,
		controller,
		in.GetFullName(),
		in.GetOwnerAnyAddress(),
		in.GetSpaceId())

	if err != nil {
		log.Fatal("can not Commit tx", zap.Error(err))
		return err
	}

	log.Info("register tx sent. Waiting for it to be mined",
		zap.String("TX hash", tx.Hash().Hex()))

	// 5 - wait for tx to be mined
	arpc.contracts.WaitMined(ctx, conn, tx)

	// 6 - return results
	txRes = arpc.contracts.CheckTransactionReceipt(conn, tx.Hash())
	if !txRes {
		// new error
		return errors.New("register tx failed")
	}

	log.Info("operation succeeded!")
	return nil
}
