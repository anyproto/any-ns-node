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

// Nonce policy:
// 1. if nonce is specified in the config file:
// - read it from config and override the value from DB/network
//
// 2. if nonce is in DB:
// - get nonce from DB
//
// 3. if nonce is not in DB:
// - get nonce from network - mined + pending txs count
//
// if tx is sent and mined succesfully:
// - save last nonce to DB
//
// if we got "nonce is too low" error the tx is immediately rejected. to fix it:
// - get nonce from network
// - send this tx again with +1 nonce
//
// if nonce is higher than needed - tx will be rejected by the network
//
// fix usually requires a manual intervention, but we implement a retry policy in case tx is too old!
// (not implemented here)
// if pending tx is too old (>N minutes), we:
// 1. get new nonce from the network
// 2. increase gas price by 10%
// 3. resend the same tx with new nonce
type NonceService interface {
	// try to determine nonce by looking in DB first, then use network as a fallback
	GetCurrentNonce(addr ethcommon.Address) (uint64, error)

	// try to determine nonce by looking at current TX count plus pending TXs in the mem pool
	// (not reliable, but can be used as a fallback)
	GetCurrentNonceFromNetwork(addr ethcommon.Address) (uint64, error)

	// save nonce to DB
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

func (anonce *anynsNonceService) GetCurrentNonce(addr ethcommon.Address) (uint64, error) {
	// 1 - if nonce is specified in the config file:
	// - read it from config and override the value from DB/network
	if anonce.confNonce.NonceOverride > 0 {
		// TODO: can not specify 0 param in config, but not a problem yet
		return anonce.confNonce.NonceOverride, nil
	}

	itemOut := &NonceDbItem{}

	// 2 - if nonce is in DB:
	// - get nonce from DB
	ctx := context.Background()
	err := anonce.nonceColl.FindOne(ctx, findNonceByAddress{Address: addr.Hex()}).Decode(&itemOut)
	if err == nil {
		// Warning: convert int64 -> uint64
		return uint64(itemOut.Nonce), nil
	}

	// 3 - if nonce is not in DB:
	return anonce.GetCurrentNonceFromNetwork(addr)
}

func (anonce *anynsNonceService) GetCurrentNonceFromNetwork(addr ethcommon.Address) (uint64, error) {
	// - get nonce from network - mined + pending txs count
	conn, err := anonce.contracts.CreateEthConnection()
	if err != nil {
		log.Error("can not create eth connection", zap.Error(err))
		return 0, err
	}

	// 2 - get gas costs, etc
	_, nonce, err := anonce.contracts.CalculateTxParams(conn, addr)
	if err != nil {
		log.Error("can not get nonce", zap.Error(err))
		return 0, err
	}
	return nonce, nil
}

// call this method when tx is sent and mined succesfully
func (anonce *anynsNonceService) SaveNonce(addr ethcommon.Address, newValue uint64) (uint64, error) {
	ctx := context.Background()
	optns := options.Replace().SetUpsert(true)

	dbItem := &NonceDbItem{
		Address: addr.Hex(),
		// TODO: conversion
		Nonce: int64(newValue),
	}

	_, err := anonce.nonceColl.ReplaceOne(ctx, findNonceByAddress{Address: addr.Hex()}, dbItem, optns)
	if err != nil {
		log.Error("failed to update item in DB", zap.Error(err))
		return 0, err
	}

	return newValue, nil
}

func (anonce *anynsNonceService) getAdminAddr() string {
	return anonce.confContracts.AddrAdmin
}
