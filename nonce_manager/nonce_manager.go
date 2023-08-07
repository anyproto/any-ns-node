package nonce_manager

import (
	"context"
	"errors"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const CName = "any-ns.nonce-manager"

var log = logger.NewNamed(CName)

// TODO: index it
type NonceDbItem struct {
	Address string `bson:"address"`
	Nonce   int64  `bson:"nonce"`
}

type findNonceByAddress struct {
	Address string `bson:"address"`
}

func New() app.Component {
	return &anynsNonceService{}
}

// NoncePolicy:
//
// if nonce is in DB:
// - get nonce from DB
//
// if nonce is not in DB:
// - get nonce from network - mined + pending txs count
//
// if nonce is specified in the config file:
// - read it from config and override the value from DB/network
//
// if tx is sent and mined succesfully:
// - save last nonce to DB
//
// if we got "nonce is too low" error the tx is immediately rejected. to fix it:
// - get nonce from network
// - send this tx again with +1 nonce
//
// if nonce is higher than needed - tx will stuck in the memory pool in pending state until:
// * other tx fill the gap
// * node is restarted
//
// fix usually requires a manual intervention, but we implement a retry policy in case tx is too OLD!
// (not implement here)
// if pending tx is too old (>N minutes), we:
// 1. get new nonce from the network
// 2. increase gas price by 10%
// 3. resend the same tx with new nonce
type NonceService interface {
	GetCurrentNonce(addr ethcommon.Address) (uint64, error)
	GetCurrentNonceFromNetwork(addr ethcommon.Address) (uint64, error)

	SaveNonce(addr ethcommon.Address, newValue uint64) (uint64, error)

	app.Component
}

type anynsNonceService struct {
	confMongo     config.Mongo
	confNonce     config.Nonce
	confContracts config.Contracts

	nonceColl *mongo.Collection
	contracts contracts.ContractsService

	//mu    sync.Mutex
}

func (anonce *anynsNonceService) Name() (name string) {
	return CName
}

func (anonce *anynsNonceService) Init(a *app.App) (err error) {
	anonce.confMongo = a.MustComponent(config.CName).(*config.Config).Mongo
	anonce.confNonce = a.MustComponent(config.CName).(*config.Config).Nonce
	anonce.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()

	anonce.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)

	uri := anonce.confMongo.Connect
	dbName := anonce.confMongo.Database
	collectionName := "nonce" // hard-coded collection

	// 1 - connect to DB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	anonce.nonceColl = client.Database(dbName).Collection(collectionName)
	if anonce.nonceColl == nil {
		return errors.New("failed to connect to MongoDB")
	}

	log.Info("nonce manager - mongo connected!")
	return nil
}

func (anonce *anynsNonceService) GetCurrentNonce() (uint64, error) {
	// 1 - if nonce is specified in the config file:
	// - read it from config and override the value from DB/network
	if anonce.confNonce.NonceOverride > 0 {
		// TODO: can not specify 0 param in config, but not a problem yet
		return anonce.confNonce.NonceOverride, nil
	}

	// 2 - if nonce is in DB:
	// - get nonce from DB
	var itemOut NonceDbItem
	adminAddr := anonce.getAdminAddr()

	ctx := context.Background()
	err := anonce.nonceColl.FindOne(ctx, findNonceByAddress{Address: adminAddr}).Decode(&itemOut)
	if err == nil {
		// Warning: convert int64 -> uint64
		return uint64(itemOut.Nonce), nil
	}

	// 3 - if nonce is not in DB:
	return anonce.GetCurrentNonceFromNetwork()
}

func (anonce *anynsNonceService) GetCurrentNonceFromNetwork() (uint64, error) {
	adminAddr := anonce.getAdminAddr()

	// - get nonce from network - mined + pending txs count
	fromAddress := ethcommon.HexToAddress(adminAddr)
	conn, err := anonce.contracts.CreateEthConnection()
	if err != nil {
		log.Error("can not create eth connection", zap.Error(err))
		return 0, err
	}

	// 2 - get gas costs, etc
	_, nonce, err := anonce.contracts.CalculateTxParams(conn, fromAddress)
	if err != nil {
		log.Error("can not get nonce", zap.Error(err))
		return 0, err
	}
	return nonce, nil
}

// call this method when tx is sent and mined succesfully
func (anonce *anynsNonceService) SaveNonce(newValue uint64) (uint64, error) {
	ctx := context.Background()

	adminAddr := anonce.getAdminAddr()
	optns := options.Replace().SetUpsert(true)

	dbItem := &NonceDbItem{
		Address: adminAddr,
		// TODO: conversion
		Nonce: int64(newValue),
	}

	_, err := anonce.nonceColl.ReplaceOne(ctx, findNonceByAddress{Address: adminAddr}, dbItem, optns)
	if err != nil {
		log.Error("failed to update item in DB", zap.Error(err))
		return 0, err
	}

	return newValue, nil
}

func (anonce *anynsNonceService) getAdminAddr() string {
	return anonce.confContracts.AddrAdmin
}
