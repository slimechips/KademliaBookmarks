package main

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

// func NewItemALT(value string) *Item {
// 	return &Item{
// 		Value:   value,
// 		Expiry:  time.Now().Add(EXPIRY_DURATION),
// 		Publish: time.Now().Add(REPUBLISHED_INITIAL_DURATION)}
// }

func (item *Item) IsTimeToPublish(current time.Time) bool {
	return current.After(item.Publish)
}

func (item *Item) IsItExpired(current time.Time) bool {
	return current.After(item.Expiry)
}

func (item *Item) String() string {
	return fmt.Sprintf("Value: %s , Publish: %s , Expiry: %s \n", item.Value, item.Publish.Local().String(), item.Expiry.Local().String())
}

// channel will alert the node every hour to see if its time to republish
func RepublishMessageNewsFlash(period chan bool) {
	time.Sleep(REPUBLISHED_DURATION)
	period <- true
}

//channel will alert node every day to check if there are data to remove
func DeleteDataIfExpireNewsFlash(period chan bool) {
	time.Sleep(EXPIRY_DURATION)
	period <- true
}
