package node

import (
	"fmt"
	"time"
)

type Item struct {
	Value   string
	Expiry  time.Time
	Publish time.Time
}

func NewItem(value string) *Item {
	return &Item{
		Value:   value,
		Expiry:  time.Now().Add(EXPIRY_DURATION),
		Publish: time.Now().Add(REPUBLISHED_DURATION)}
}

func NewItemALT(value string) *Item {
	return &Item{
		Value:   value,
		Expiry:  time.Now().Add(EXPIRY_DURATION),
		Publish: time.Now().Add(REPUBLISHED_INITIAL_DURATION)}
}

func (item *Item) IsTimeToPublish(current time.Time) bool {
	return current.After(item.Publish)
}

func (item *Item) String() string {
	return fmt.Sprintf("Value: %s , Publish: %s , Expiry: %s \n", item.Value, item.Publish.Local().String(), item.Expiry.Local().String())
}
