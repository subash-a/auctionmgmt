package main

import (
	"testing"
	"time"
)

func TestAddBid(t *testing.T) {
	start := time.Date(2019, time.November, 18, 0, 0, 0, 0, time.Local)
	end := time.Date(2019, time.November, 19, 0, 0, 0, 0, time.Local)
	price := float32(20.45)

	// correct call should not fail
	_, err := NewAuction(start, end, price)
	if err != nil {
		t.Fatalf("should not have seen error but saw %v\n", err)
	}

	// testing invalid start
	_, err = NewAuction(end, start, price)
	if err == nil {
		t.Fatalf("should have seen invlaid start date error but did not")
	}

	// testing invalid price
	_, err = NewAuction(start, end, -price)
	if err == nil {
		t.Fatalf("should have seen invlaid price error but did not")
	}
}

func TestGetResult(t *testing.T) {
	start := time.Date(2019, time.November, 12, 0, 0, 0, 0, time.Local)
	end := time.Date(2019, time.November, 13, 0, 0, 0, 0, time.Local)
	price := float32(20.45)

	bid0 := NewBid("01", float32(21.45))
	bid1 := NewBid("02", float32(21.65))
	bid2 := NewBid("03", float32(21.85))

	bids := []bid{bid0, bid1, bid2}
	a := &auction{
		Id:        generateId(),
		Start:     start,
		End:       end,
		Bids:      bids,
		InitPrice: price,
	}

	res, err := a.GetResult()
	if err != nil {
		t.Fatalf("should not have errored but did %v\n", err)
	}

	if res != "03" {
		t.Fatalf("error in processing winning bid")
	}
}
