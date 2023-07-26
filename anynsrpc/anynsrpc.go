package anynsrpc

import (
	"context"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-cid"
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
		log.Error("failed to connect to geth", zap.Error(err))
		return nil, err
	}

	// 1 - convert to name hash
	nh, err := contracts.NameHash(in.FullName)
	if err != nil {
		log.Error("can not convert FullName to namehash", zap.Error(err))
		return nil, err
	}

	// 2 - call contract's method
	log.Info("getting owner for name", zap.String("FullName", in.GetFullName()))
	addr, err := arpc.contracts.GetOwnerForNamehash(conn, nh)
	if (err != nil) || (addr == nil) {
		log.Error("failed to get owner", zap.Error(err))
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
		log.Error("failed to get additional info", zap.Error(err))
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
		log.Error("can not register name", zap.Error(err))
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
		log.Error("can not unmarshal NameRegisterRequest", zap.Error(err))
		return &resp, err
	}

	// 2 - check signature
	err = VerifyIdentity(in, nrr.OwnerAnyAddress)
	if err != nil {
		resp.OperationState = as.OperationState_Error
		log.Error("identity is different", zap.Error(err))
		return &resp, err
	}

	// 3 - finally call function
	err = arpc.nameRegister(ctx, &nrr)

	if err != nil {
		log.Error("can not register name", zap.Error(err))
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

	// 0 - check all parameters
	err := arpc.checkRegisterParams(in)
	if err != nil {
		log.Error("invalid parameters", zap.Error(err))
		return err
	}

	conn, err := arpc.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return err
	}

	// 1 - connect to geth
	controller, err := arpc.contracts.ConnectToController(conn)
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return err
	}

	// 2 - get a name's first part
	// TODO: normalize string
	nameFirstPart := contracts.RemoveTLD(in.FullName)

	// 3 - calculate a commitment
	secret, err := contracts.GenerateRandomSecret()
	if err != nil {
		log.Error("can not generate random secret", zap.Error(err))
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
		log.Error("can not calculate a commitment", zap.Error(err))
		return err
	}

	authOpts, err := arpc.contracts.GenerateAuthOptsForAdmin(conn)
	if err != nil {
		log.Error("can not get auth params for admin", zap.Error(err))
		return err
	}

	// 4 - commit from Admin
	tx, err := arpc.contracts.Commit(
		authOpts,
		commitment,
		controller)

	// TODO: check if tx is nil?
	if err != nil {
		log.Error("can not Commit tx", zap.Error(err))
		return err
	}

	// wait for tx to be mined
	txRes, err := arpc.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for commit tx", zap.Error(err))
		return err
	}
	if !txRes {
		// new error
		return errors.New("commit tx not mined")
	}

	// update nonce again...
	authOpts, err = arpc.contracts.GenerateAuthOptsForAdmin(conn)
	if err != nil {
		log.Error("can not get auth params for admin", zap.Error(err))
		return err
	}

	// 5 - register
	tx, err = arpc.contracts.Register(
		authOpts,
		nameFirstPart,
		registrantAccount,
		secret,
		controller,
		in.GetFullName(),
		in.GetOwnerAnyAddress(),
		in.GetSpaceId())

	// TODO: check if tx is nil?
	if err != nil {
		log.Error("can not Commit tx", zap.Error(err))
		return err
	}

	// wait for tx to be mined
	txRes, err = arpc.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for register tx", zap.Error(err))
		return err
	}
	if !txRes {
		// new error
		return errors.New("register tx failed")
	}

	log.Info("operation succeeded!")
	return nil
}

func (arpc *anynsRpc) checkRegisterParams(in *as.NameRegisterRequest) error {
	// 1 - check name
	if !arpc.checkName(in.FullName) {
		log.Error("invalid name", zap.String("name", in.FullName))
		return errors.New("invalid name")
	}

	// 2 - check ETH address
	if !common.IsHexAddress(in.OwnerEthAddress) {
		log.Error("invalid ETH address", zap.String("ETH address", in.OwnerEthAddress))
		return errors.New("invalid ETH address")
	}

	// 3 - check Any address
	if !arpc.checkAnyAddress(in.OwnerAnyAddress) {
		log.Error("invalid Any address", zap.String("Any address", in.OwnerAnyAddress))
		return errors.New("invalid Any address")
	}

	// 4 - space ID (if not empty)
	if in.SpaceId != "" {
		_, err := cid.Decode(in.SpaceId)

		if err != nil {
			log.Error("invalid SpaceId", zap.String("Any SpaceId", in.SpaceId))
			return errors.New("invalid SpaceId")
		}
	}

	// everything is OK
	return nil
}

func (arpc *anynsRpc) checkName(name string) bool {
	// get name parts
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return false
	}

	// if extension is not 'any', then return false
	if parts[len(parts)-1] != "any" {
		return false
	}

	// if first part is less than 3 chars, then return false
	if len(parts[0]) < 3 {
		return false
	}

	// if it has slashes, then return false
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}

	return true
}

func isValidAnyAddress(address string) bool {
	// correct address format is 12D3KooWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy
	// it should start with 1
	if !strings.HasPrefix(address, "1") {
		return false
	}

	// the len should be 52
	if len(address) != 52 {
		return false
	}

	return true
}

func (arpc *anynsRpc) checkAnyAddress(addr string) bool {
	// in.OwnerAnyAddress should be a ed25519 public key hash
	return isValidAnyAddress(addr)
}
