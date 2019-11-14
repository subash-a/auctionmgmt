package main

import (
	"fmt"
	"time"
)

type auction struct {
	Id        string
	Start     time.Time
	End       time.Time
	InitPrice float32
	Bids      []bid
}

func NewAuction(key string, start, end time.Time, price float32) (*auction, error) {
	return &auction{
		Id:        "abc",
		Start:     start,
		End:       end,
		InitPrice: price,
		Bids:      []bid{},
	}, nil
}

func (a *auction) AddBid(b bid) error {
	if !a.InProgress() {
		return fmt.Errorf("auction is closed cannot bid")
	}

	a.Bids = append(a.Bids, b)
	return nil
}

func (a *auction) InProgress() bool {
	now := time.Now()
	if now.After(a.Start) && now.Before(a.End) {
		return true
	}

	return false
}

func keyAuthentic(key string) bool {
	// add logic here to verify secret key and return response
	return true
}

func (a *auction) GetResult() (string, error) {
	if a.InProgress() {
		return "", fmt.Errorf("auction still in progress")
	}

	if len(a.Bids) == 0 {
		return "", fmt.Errorf("no bids for auction, no winner")
	}

	winningBid := a.Bids[0]
	for _, b := range a.Bids {
		if b.Price > winningBid.Price {
			winningBid = b
		}
	}

	return winningBid.UserId, nil
}

type bid struct {
	UserId string
	Price  float32
}

func NewBid(id string, price float32) bid {
	return bid{
		UserId: id,
		Price:  price,
	}
}

type fakeDB struct {
	auctions map[string]*auction
}

func NewFakeDB() *fakeDB {
	return &fakeDB{auctions: map[string]*auction{}}
}

func (f *fakeDB) AllActiveAuctions() []*auction {
	res := []*auction{}

	for _, a := range f.auctions {
		if a.InProgress() {
			res = append(res, a)
		}
	}

	return res
}

func (f *fakeDB) AllCompletedAuctions() []*auction {
	res := []*auction{}
	currtime := time.Now()
	for _, a := range f.auctions {
		if currtime.After(a.End) {
			res = append(res, a)
		}
	}

	return res
}

func (f *fakeDB) AllPendingAuctions() []*auction {
	res := []*auction{}
	currtime := time.Now()
	for _, a := range f.auctions {
		if currtime.Before(a.Start) {
			res = append(res, a)
		}
	}

	return res
}

// CRUD operations authenticated by secret key
func (f *fakeDB) Add(key string, auction *auction) error {
	// authenticate key and then return a new auction instance
	if !keyAuthentic(key) {
		return fmt.Errorf("invalid secret key")
	}

	if f.auctions[auction.Id] != nil {
		return fmt.Errorf("entry already exists")
	}

	f.auctions[auction.Id] = auction
	return nil
}

func (f *fakeDB) Delete(key string, id string) error {
	// authenticate key and then return a new auction instance
	if !keyAuthentic(key) {
		return fmt.Errorf("invalid secret key")
	}
	f.auctions[id] = nil
	return nil
}

func (f *fakeDB) Update(key string, id string, auction *auction) error {
	// authenticate key and then return a new auction instance
	if !keyAuthentic(key) {
		return fmt.Errorf("invalid secret key")
	}
	f.auctions[id] = auction
	return nil
}

func (f *fakeDB) Get(key string, id string) (*auction, error) {
	// authenticate key and then return a new auction instance
	if !keyAuthentic(key) {
		return nil, fmt.Errorf("invalid secret key")
	}

	return f.auctions[id], nil
}

func main() {
	db := NewFakeDB()
}
