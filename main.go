package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type auction struct {
	Id        string
	Start     time.Time
	End       time.Time
	InitPrice float32
	Bids      []bid
}

func NewAuction(start, end time.Time, price float32) (*auction, error) {
	if start.After(end) {
		return nil, fmt.Errorf("start date is invalid")
	}

	if price < 0 {
		return nil, fmt.Errorf("price cannot be negative")
	}

	return &auction{
		Id:        generateId(),
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

	if b.Price < a.InitPrice {
		return fmt.Errorf("invalid price for bid")
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
	UserId    string
	Price     float32
	Timestamp time.Time // used for identifying the earliest bid when there is a tie between price
}

func NewBid(id string, price float32) bid {
	return bid{
		UserId:    id,
		Price:     price,
		Timestamp: time.Now(),
	}
}

// NOTE: this is DB equivalent here which stores the data
// until the program terminates
type persistence struct {
	auctions map[string]*auction
}

func NewPersistence() *persistence {
	return &persistence{auctions: map[string]*auction{}}
}

func (f *persistence) AllActiveAuctions() []*auction {
	res := []*auction{}

	for _, a := range f.auctions {
		if a.InProgress() {
			res = append(res, a)
		}
	}

	return res
}

func (f *persistence) AllCompletedAuctions() []*auction {
	res := []*auction{}
	currtime := time.Now()
	for _, a := range f.auctions {
		if currtime.After(a.End) {
			res = append(res, a)
		}
	}

	return res
}

func (f *persistence) AllPendingAuctions() []*auction {
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
func (f *persistence) Add(key string, auction *auction) error {
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

func (f *persistence) Delete(key string, id string) error {
	// authenticate key and then return a new auction instance
	if !keyAuthentic(key) {
		return fmt.Errorf("invalid secret key")
	}
	f.auctions[id] = nil
	return nil
}

func (f *persistence) Update(key string, id string, auction *auction) error {
	// updating an auction might not be the right thing to do specially if the
	// initial price, start time, etc is being changed because this may lead to
	// inconsistencies in the bidding process as some bids may not satisfy some
	// of the criteria like price hence we should disallow updating after the auction
	// has commenced

	// authenticate key and then return a new auction instance
	if !keyAuthentic(key) {
		return fmt.Errorf("invalid secret key")
	}

	oldAuction := f.auctions[id]
	if oldAuction.InProgress() {
		return fmt.Errorf("auction in progress, cannot be updated")
	}

	auction.Bids = oldAuction.Bids
	f.auctions[id] = auction
	return nil
}

func (f *persistence) Get(key string, id string) (*auction, error) {
	// authenticate key and then return a new auction instance
	if !keyAuthentic(key) {
		return nil, fmt.Errorf("invalid secret key")
	}

	return f.auctions[id], nil
}

func generateId() string {
	id := uuid.New() // generates a new random UUID
	return id.String()
}

func main() {
	if err := mainerr(); err != nil {
		panic(err)
	}
}

func mainerr() error {
	var e error

	db := NewPersistence()

	mux := http.NewServeMux()
	mux.HandleFunc("/app/login", func(w http.ResponseWriter, req *http.Request) {
		// handle authentication request for login
		respObj := struct {
			Key   string `json:"key,omitempty"`
			Error string `json:"error,omitempty"`
		}{Key: "secret-key", Error: ""}
		b, err := json.Marshal(respObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}
		// add the below headers to allow CORS since we are using two different ports
		// one to serve the backend and one to serve the frontend
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Add("Access-Control-Allow-Credentials", "true")

		w.Header().Add("Content-Type", "application/json")
		_, err = w.Write(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/app/auctions", func(w http.ResponseWriter, req *http.Request) {
		// handle authentication request for login
		respObj := struct {
			Auctions []*auction `json:"auctions,omitempty"`
			Error    string     `json:"error,omitempty"`
		}{Auctions: db.AllActiveAuctions(), Error: ""}
		b, err := json.Marshal(respObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}
		// add the below headers to allow CORS since we are using two different ports
		// one to serve the backend and one to serve the frontend
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Add("Access-Control-Allow-Credentials", "true")

		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	})

	mux.HandleFunc("/app/auctions/add", func(w http.ResponseWriter, req *http.Request) {
		// handle authentication request for login
		body, err := req.GetBody()
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		content, err := ioutil.ReadAll(body)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		// marshal into JSON and then read key and auction id
		var reqObj struct {
			Auction *auction `json:"auction,omitempty"`
			Key     string   `json:"key,omitempty"`
		}
		err = json.Unmarshal(content, &reqObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		err = db.Add(reqObj.Key, reqObj.Auction)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		// resp is a response object which contains the query response
		respObj := struct {
			Auctions []*auction `json:"auctions,omitempty"`
			Error    string     `json:"error,omitempty"`
		}{Auctions: db.AllActiveAuctions(), Error: err.Error()}
		b, err := json.Marshal(respObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}
		// add the below headers to allow CORS since we are using two different ports
		// one to serve the backend and one to serve the frontend
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Add("Access-Control-Allow-Credentials", "true")

		w.Header().Add("Content-Type", "application/json")
		w.Write(b)

	})

	mux.HandleFunc("/app/auctions/update", func(w http.ResponseWriter, req *http.Request) {
		// handle authentication request for login
		body, err := req.GetBody()
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		content, err := ioutil.ReadAll(body)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		// marshal into JSON and then read key and auction id
		var reqObj struct {
			Auction *auction `json:"auction,omitempty"`
			Key     string   `json:"key,omitempty"`
		}
		err = json.Unmarshal(content, &reqObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		err = db.Add(reqObj.Key, reqObj.Auction)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		// resp is a response object which contains the query response
		respObj := struct {
			Auctions []*auction `json:"auctions,omitempty"`
			Error    string     `json:"error,omitempty"`
		}{Auctions: db.AllActiveAuctions(), Error: err.Error()}
		b, err := json.Marshal(respObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}
		// add the below headers to allow CORS since we are using two different ports
		// one to serve the backend and one to serve the frontend
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Add("Access-Control-Allow-Credentials", "true")

		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
		// handle updating an auction
	})

	mux.HandleFunc("/app/auctions/get", func(w http.ResponseWriter, req *http.Request) {
		// handle fetching an auction status
		body, err := req.GetBody()
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		content, err := ioutil.ReadAll(body)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		// marshal into JSON and then read key and auction id
		var reqObj struct {
			Id  string `json:"id,omitempty"`
			Key string `json:"key,omitempty"`
		}
		err = json.Unmarshal(content, &reqObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		a, err := db.Get(reqObj.Key, reqObj.Id)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		// resp is a response object which contains the query response
		respObj := struct {
			Auctions []*auction `json:"auctions,omitempty"`
			Error    string     `json:"error,omitempty"`
		}{Auctions: []*auction{a}, Error: err.Error()}
		b, err := json.Marshal(respObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}
		// add the below headers to allow CORS since we are using two different ports
		// one to serve the backend and one to serve the frontend
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Add("Access-Control-Allow-Credentials", "true")

		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	})

	mux.HandleFunc("/app/auctions/delete", func(w http.ResponseWriter, req *http.Request) {
		// handle removal of auction
		body, err := req.GetBody()
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		content, err := ioutil.ReadAll(body)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		// marshal into JSON and then read key and auction id
		var reqObj struct {
			Auction *auction `json:"auction,omitempty"`
			Key     string   `json:"key,omitempty"`
		}
		err = json.Unmarshal(content, &reqObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		err = db.Add(reqObj.Key, reqObj.Auction)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		// resp is a response object which contains the query response
		respObj := struct {
			Auctions []*auction `json:"auctions,omitempty"`
			Error    string     `json:"error,omitempty"`
		}{Auctions: db.AllActiveAuctions(), Error: err.Error()}
		b, err := json.Marshal(respObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}
		// add the below headers to allow CORS since we are using two different ports
		// one to serve the backend and one to serve the frontend
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Add("Access-Control-Allow-Credentials", "true")

		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	})

	// handle normal user requests
	mux.HandleFunc("/app/auctions/bid", func(w http.ResponseWriter, req *http.Request) {
		// handle bidding for an auction
		body, err := req.GetBody()
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		content, err := ioutil.ReadAll(body)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		// marshal into JSON and then read key and auction id
		var reqObj struct {
			Id  string `json:"id,omitempty"`
			Bid bid    `json:"bid,omitempty"`
		}
		err = json.Unmarshal(content, &reqObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}

		auctions := db.AllActiveAuctions()
		var atn *auction
		for _, a := range auctions {
			if a.Id == reqObj.Id {
				atn = a
			}
		}

		// NOTE: we can add a check here to see if atn is nil before adding the bid
		// just in case we encounter an invalid auction id, but this is not done
		// here since handling this validation in the frontend would be more desirable
		// as handling for every kind of scenario on the backend would make backend code
		// very complex and difficult to maintain

		err = atn.AddBid(reqObj.Bid)
		// resp is a response object which contains the query response
		respObj := struct {
			Auctions []*auction `json:"auctions,omitempty"`
			Error    string     `json:"error,omitempty"`
		}{Auctions: db.AllActiveAuctions(), Error: err.Error()}
		b, err := json.Marshal(respObj)
		if err != nil {
			e = err
			w.WriteHeader(http.StatusInternalServerError)
		}
		// add the below headers to allow CORS since we are using two different ports
		// one to serve the backend and one to serve the frontend
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Add("Access-Control-Allow-Credentials", "true")

		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	})

	srv := http.Server{
		Addr:    ":7070",
		Handler: mux,
	}

	fmt.Println("Running Server on port 7070...")
	err := srv.ListenAndServe()
	if err != nil {
		e = err
		srv.Close()
	}

	return e
}
