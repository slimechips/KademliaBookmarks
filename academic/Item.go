package academic

import "time"

type Item struct {
	Key     ID
	Value   string
	Expiry  time.Time
	Publish time.Time
}

func NewItem(key ID, value string) *Item {
	return &Item{
		Key:     key,
		Value:   value,
		Expiry:  time.Now().Add(EXPIRY_DURATION),
		Publish: time.Now().Add(REPUBLISHED_DURATION)}
}

func (item *Item) IsTimeToPublish(current time.Time) bool {
	return current.After(item)
}
