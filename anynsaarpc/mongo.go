package anynsaarpc

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func (arpc *anynsAARpc) mongoAddUserToTheWhitelist(ctx context.Context, owner common.Address, ownerAnyID string, newOperations uint64) (err error) {
	// TODO: rewrite atomically

	// 1 - verify parameters
	if newOperations == 0 {
		return errors.New("wrong operations count")
	}

	// 2 - get item from mongo
	item := &AAUser{}
	err = arpc.itemColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)

	if err != nil {
		// 3.1 - if not found - create new
		if err == mongo.ErrNoDocuments {
			_, err = arpc.itemColl.InsertOne(ctx, AAUser{
				Address:         owner.Hex(),
				AnyID:           ownerAnyID,
				OperationsCount: newOperations,
			})
			if err != nil {
				log.Error("failed to insert item to DB", zap.Error(err))
				return err
			}
			log.Info("added new user to the whitelist", zap.String("owner", owner.Hex()))
			return nil
		}

		log.Error("failed to get item from DB", zap.Error(err))
		return err
	}

	// 3.2 - update item in mongo
	// but first check if Any ID is the same as was passed above
	if ownerAnyID != item.AnyID {
		log.Error("AnyID does not match", zap.String("any_id", ownerAnyID), zap.String("item.AnyID", item.AnyID))
		return errors.New("AnyID does not match")
	}

	log.Debug("increasing operations count in the whitelist", zap.String("owner", owner.Hex()))

	optns := options.Replace().SetUpsert(true)
	item.OperationsCount += newOperations // update operations count

	// 4 - write it back to DB
	_, err = arpc.itemColl.ReplaceOne(ctx, findAAUserByAddress{Address: owner.Hex()}, item, optns)
	if err != nil {
		log.Error("failed to update item in DB", zap.Error(err))
		return err
	}

	log.Info("updated whitelist", zap.String("owner", owner.Hex()))
	return nil
}

// will check if ownerAnyID matches AnyID in the DB (was set by Admin before)
// if ownerAnyID is empty -> do not check it
func (arpc *anynsAARpc) mongoGetUserOperationsCount(ctx context.Context, owner common.Address, ownerAnyID string) (operations uint64, err error) {
	item := &AAUser{}
	err = arpc.itemColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)

	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return 0, err
	}

	// check if AnyID is correct
	// this should be in the format of PeerID - 12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS
	if (ownerAnyID != "") && (item.AnyID != ownerAnyID) {
		log.Error("AnyID does not match", zap.String("any_id", ownerAnyID))
		return 0, errors.New("AnyID does not match")
	}

	return item.OperationsCount, nil
}

func (arpc *anynsAARpc) mongoDecreaseUserOperationsCount(ctx context.Context, owner common.Address) (err error) {
	// TODO: rewrite atomically

	// 1 - get item from mongo
	item := &AAUser{}
	err = arpc.itemColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)

	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return err
	}
	if item.OperationsCount == 0 {
		log.Error("operations count is already 0", zap.String("owner", owner.Hex()))
		return errors.New("operations count is already 0")
	}

	// 2 - update item in mongo
	log.Debug("decreasing operations count in the whitelist", zap.String("owner", owner.Hex()))

	optns := options.Replace().SetUpsert(false)
	// update operations count
	item.OperationsCount -= 1

	_, err = arpc.itemColl.ReplaceOne(ctx, findAAUserByAddress{Address: owner.Hex()}, item, optns)
	if err != nil {
		log.Error("failed to update item in DB", zap.Error(err))
		return err
	}

	log.Info("decreased op count in the whitelist", zap.String("owner", owner.Hex()))
	return nil
}
