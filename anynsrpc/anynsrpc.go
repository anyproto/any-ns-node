package anynsrpc

import (
	"context"
	"errors"
	"slices"

	"github.com/anyproto/any-ns-node/cache"
	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/queue"

	"github.com/anyproto/any-ns-node/verification"
	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/peer"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/anyproto/any-sync/nodeconf"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	contracts "github.com/anyproto/any-ns-node/contracts"
	dbservice "github.com/anyproto/any-ns-node/db"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

const CName = "any-ns.rpc"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsRpc{}
}

type anynsRpc struct {
	cache         cache.CacheService
	conf          *config.Config
	confContracts config.Contracts
	confAccount   accountservice.Config
	nodeConf      nodeconf.NodeConf
	contracts     contracts.ContractsService
	queue         queue.QueueService
	aa            accountabstraction.AccountAbstractionService
	db            dbservice.DbService

	readFromCache bool
}

func (arpc *anynsRpc) Init(a *app.App) (err error) {
	arpc.cache = a.MustComponent(cache.CName).(cache.CacheService)
	arpc.conf = a.MustComponent(config.CName).(*config.Config)
	arpc.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.confAccount = a.MustComponent(config.CName).(*config.Config).GetAccount()
	arpc.nodeConf = a.MustComponent(nodeconf.CName).(nodeconf.NodeConf)
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	arpc.readFromCache = a.MustComponent(config.CName).(*config.Config).ReadFromCache
	arpc.queue = a.MustComponent(queue.CName).(queue.QueueService)
	arpc.aa = a.MustComponent(accountabstraction.CName).(accountabstraction.AccountAbstractionService)
	arpc.db = a.MustComponent(dbservice.CName).(dbservice.DbService)

	return nsp.DRPCRegisterAnyns(a.MustComponent(server.CName).(server.DRPCServer), arpc)
}

func (arpc *anynsRpc) Name() (name string) {
	return CName
}

func (arpc *anynsRpc) IsNameAvailable(ctx context.Context, in *nsp.NameAvailableRequest) (*nsp.NameAvailableResponse, error) {
	// 0 - normalize name (including .any suffix)
	useEnsip15 := arpc.conf.Ensip15Validation
	fullName, err := contracts.NormalizeAnyName(in.FullName, useEnsip15)
	if err != nil {
		log.Error("failed to normalize name", zap.Error(err))
		return nil, err
	}

	in.FullName = fullName

	// 1 - if ReadFromCache is false -> always first read from smart contracts
	// if not, then always just read quickly from cache
	if !arpc.readFromCache {
		log.Debug("EXCPLICIT: read data from smart contracts -> cache", zap.String("FullName", in.FullName))
		err := arpc.cache.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: in.FullName,
		})

		if err != nil {
			log.Error("failed to update in cache", zap.Error(err))
			return nil, errors.New("failed to update in cache")
		}
	}

	// 2 - check in cache (Mongo)
	return arpc.cache.IsNameAvailable(ctx, in)
}

func (arpc *anynsRpc) GetNameByAddress(ctx context.Context, in *nsp.NameByAddressRequest) (*nsp.NameByAddressResponse, error) {
	// 1 - if ReadFromCache is false -> always first read from smart contracts
	// if not, then always just read quickly from cache
	if !arpc.readFromCache {
		log.Debug("EXCPLICIT: reverse resolve using no cache", zap.String("FullName", in.OwnerScwEthAddress))
		return arpc.getNameByAddressDirectly(ctx, in)
	}

	// check in cache (Mongo)
	return arpc.cache.GetNameByAddress(ctx, in)
}

func (arpc *anynsRpc) GetNameByAnyId(ctx context.Context, in *nsp.NameByAnyIdRequest) (*nsp.NameByAddressResponse, error) {
	// this method always reads from cache!
	// there is no way to directly do reverse resolve using smart contracts
	// (for now)
	return arpc.cache.GetNameByAnyId(ctx, in)
}

func (arpc *anynsRpc) getNameByAddressDirectly(ctx context.Context, in *nsp.NameByAddressRequest) (*nsp.NameByAddressResponse, error) {
	// 1 - check parameters
	if !common.IsHexAddress(in.OwnerScwEthAddress) {
		log.Error("invalid ETH address", zap.String("ETH address", in.OwnerScwEthAddress))
		return nil, errors.New("invalid ETH address")
	}

	// convert in.OwnerScwEthAddress to common.Address
	var addr = common.HexToAddress(in.OwnerScwEthAddress)

	name, err := arpc.contracts.GetNameByAddress(addr)
	if err != nil {
		log.Error("failed to get name by address", zap.Error(err))
		return nil, errors.New("failed to get name by address")
	}

	// 2 - return results
	var res nsp.NameByAddressResponse

	if name == "" {
		res.Found = false
		return &res, nil
	}

	res.Found = true
	res.Name = name
	return &res, nil
}

func (arpc *anynsRpc) isAdmin(peerId string) bool {
	// 1 - check if peer is a payment node!
	if slices.Contains(arpc.nodeConf.NodeTypes(peerId), nodeconf.NodeTypePaymentProcessingNode) {
		return true
	}

	// 2 - admin
	err := verification.VerifyAdminIdentity(arpc.confAccount.PeerKey, peerId)
	return (err == nil)
}

func (arpc *anynsRpc) AdminNameRegisterSigned(ctx context.Context, in *nsp.NameRegisterRequestSigned) (*nsp.OperationResponse, error) {
	peerId, err := peer.CtxPeerId(ctx)
	if err != nil {
		return nil, err
	}

	var resp nsp.OperationResponse

	// 1 - unmarshal the signed request
	var nrr nsp.NameRegisterRequest
	err = nrr.UnmarshalVT(in.Payload)
	if err != nil {
		resp.OperationState = nsp.OperationState_Error
		log.Error("can not unmarshal NameRegisterRequest", zap.Error(err))
		return &resp, err
	}

	// 2 - check signature
	err = verification.VerifyAdminIdentity(arpc.confAccount.PeerKey, peerId)
	isAllow := arpc.isAdmin(peerId)
	if !isAllow {
		log.Error("not an Admin!!!", zap.Error(err))
		return nil, errors.New("not an Admin!!!")
	}

	// 3 - check all parameters
	useEnsip15 := arpc.conf.Ensip15Validation
	err = verification.CheckRegisterParams(&nrr, useEnsip15)
	if err != nil {
		log.Error("invalid parameters", zap.Error(err))
		return nil, err
	}

	// old version: process it manually in the queue
	/*
		// 4 - add to queue
		operationId, err := arpc.queue.AddNewRequest(ctx, &nrr)
		resp.OperationId = fmt.Sprint(operationId)
		resp.OperationState = nsp.OperationState_Pending
		return &resp, err
	*/

	// 4 - new version - use AA to process it
	opID, err := arpc.aa.AdminNameRegister(ctx, &nrr)
	if err != nil {
		log.Error("failed to process AdminNameRegister", zap.Error(err))
		return nil, err
	}

	// 5 - save operation to mongo (can be used later)
	cuor := nsp.CreateUserOperationRequest{
		//Data: [],
		//SignedData: [],
		//Context: [],
		OwnerAnyID:      nrr.OwnerAnyAddress,
		FullName:        nrr.FullName,
		OwnerEthAddress: nrr.OwnerEthAddress,
	}
	err = arpc.db.SaveOperation(ctx, opID, cuor)
	if err != nil {
		log.Error("failed to save operation to Mongo", zap.Error(err))
		return nil, errors.New("failed to save operation")
	}

	var out nsp.OperationResponse
	out.OperationId = opID
	out.OperationState = nsp.OperationState_Pending

	return &out, err
}

func (arpc *anynsRpc) AdminNameRenewSigned(ctx context.Context, in *nsp.NameRenewRequestSigned) (*nsp.OperationResponse, error) {
	peerId, err := peer.CtxPeerId(ctx)
	if err != nil {
		return nil, err
	}

	var resp nsp.OperationResponse

	// 1 - unmarshal the signed request
	var nrr nsp.NameRenewRequest
	err = nrr.UnmarshalVT(in.Payload)
	if err != nil {
		resp.OperationState = nsp.OperationState_Error
		log.Error("can not unmarshal NameRegisterRequest", zap.Error(err))
		return &resp, err
	}

	// 2 - check signature
	err = verification.VerifyAdminIdentity(arpc.confAccount.PeerKey, peerId)
	isAllow := arpc.isAdmin(peerId)
	if !isAllow {
		log.Error("not an Admin!!!", zap.Error(err))
		return nil, errors.New("not an Admin!!!")
	}

	// TODO: validate renew parameters without waiting for TX to fail in the smart contract

	// old version: process it manually in the queue
	/*
		// 4 - add to queue
		operationId, err := arpc.queue.AddNewRequest(ctx, &nrr)
		resp.OperationId = fmt.Sprint(operationId)
		resp.OperationState = nsp.OperationState_Pending
		return &resp, err
	*/

	// 4 - new version - use AA to process it
	opID, err := arpc.aa.AdminNameRenew(ctx, &nrr)
	if err != nil {
		log.Error("failed to process AdminNameRegister", zap.Error(err))
		return nil, err
	}

	// 5 - save operation to mongo (can be used later)
	cuor := nsp.CreateUserOperationRequest{
		//Data: [],
		//SignedData: [],
		//Context: [],
		OwnerAnyID:      nrr.OwnerAnyAddress,
		FullName:        nrr.FullName,
		OwnerEthAddress: nrr.OwnerEthAddress,
	}
	err = arpc.db.SaveOperation(ctx, opID, cuor)
	if err != nil {
		log.Error("failed to save operation to Mongo", zap.Error(err))
		return nil, errors.New("failed to save operation")
	}

	var out nsp.OperationResponse
	out.OperationId = opID
	out.OperationState = nsp.OperationState_Pending

	return &out, err
}

// Batch methods
func (arpc *anynsRpc) BatchIsNameAvailable(ctx context.Context, in *nsp.BatchNameAvailableRequest) (out *nsp.BatchNameAvailableResponse, err error) {
	// for each string in in.FullNames call IsNameAvailable and collect results into out.NameAvailableResponse[]
	out = &nsp.BatchNameAvailableResponse{
		Results: make([]*nsp.NameAvailableResponse, len(in.FullNames)),
	}

	for i, fullName := range in.FullNames {
		resp, err := arpc.IsNameAvailable(ctx, &nsp.NameAvailableRequest{
			FullName: fullName,
		})

		// do not ignore error here, stop the cycle!
		if err != nil {
			log.Error("failed to call IsNameAvailable", zap.Error(err))
			return nil, err
		}
		out.Results[i] = resp
	}

	return out, nil
}

func (arpc *anynsRpc) BatchGetNameByAddress(ctx context.Context, in *nsp.BatchNameByAddressRequest) (*nsp.BatchNameByAddressResponse, error) {
	// for each in.OwnerScwEthAddresses call GetNameByAddress and collect results into out.NameByAddressResponse[]
	out := &nsp.BatchNameByAddressResponse{
		Results: make([]*nsp.NameByAddressResponse, len(in.OwnerScwEthAddresses)),
	}

	for i, addr := range in.OwnerScwEthAddresses {
		resp, err := arpc.GetNameByAddress(ctx, &nsp.NameByAddressRequest{
			OwnerScwEthAddress: addr,
		})

		// do not ignore error here, stop the cycle!
		if err != nil {
			log.Error("failed to call GetNameByAddress", zap.Error(err))
			return nil, err
		}
		out.Results[i] = resp
	}

	return out, nil
}

func (arpc *anynsRpc) BatchGetNameByAnyId(ctx context.Context, in *nsp.BatchNameByAnyIdRequest) (*nsp.BatchNameByAddressResponse, error) {
	// for each in.AnyAddresses call GetNameByAnyId and collect results into out.NameByAddressResponse[]
	out := &nsp.BatchNameByAddressResponse{
		Results: make([]*nsp.NameByAddressResponse, len(in.AnyAddresses)),
	}

	for i, addr := range in.AnyAddresses {
		resp, err := arpc.GetNameByAnyId(ctx, &nsp.NameByAnyIdRequest{
			AnyAddress: addr,
		})

		// do not ignore error here, stop the cycle!
		if err != nil {
			log.Error("failed to call GetNameByAnyId", zap.Error(err))
			return nil, err
		}
		out.Results[i] = resp
	}

	return out, nil
}
