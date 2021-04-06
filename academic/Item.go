package academic

import "time"

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

func (item *Item) IsTimeToPublish(current time.Time) bool {
	return current.After(item)
}
