package orderhelp

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type BriefOrderInfo struct {
	Store   string `json:"s"`
	OrderID string `json:"o"`
}

func IntentToOrderGet(rdb *redis.Client, pmid string) (BriefOrderInfo, error) {
	key := pmid + "::PMID"
	val, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		return BriefOrderInfo{}, err
	}
	var orderInfo BriefOrderInfo
	err = json.Unmarshal([]byte(val), &orderInfo)
	if err != nil {
		return BriefOrderInfo{}, err
	}
	return orderInfo, nil
}

func IntentToOrderSet(rdb *redis.Client, pmid, store, orderID string) error {
	key := pmid + "::PMID"
	orderInfo := BriefOrderInfo{Store: store, OrderID: orderID}
	data, err := json.Marshal(orderInfo)
	if err != nil {
		return err
	}
	err = rdb.Set(context.Background(), key, data, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func IntentToOrderUnet(rdb *redis.Client, pmid string) error {
	key := pmid + "::PMID"
	err := rdb.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
