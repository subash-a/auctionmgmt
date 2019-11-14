## Auction Management Service
------------------------------

### Description

This is an auction and bidding management service implemented in Go. The objective of this system are as follows

* Allow administrators to setup auctions by specifying criteria like auction start time, end time and starting price for the product
* Allow users of the service to bid for all active auctions
* Users can bid any number of times for an auction
* Once the auction end time is complete the winning bid user id can be retrieved and published
* Administrators can track the status of all bids of an auction at any given point of time

### Setup
The service can be setup by following the steps below

1. Download the repository into a particular directory
2. Enter the directory and run the command `go run main.go`
3. The above command would enable the backend service to start
4. Now open another terminal window and go to the root directory of the application
5. Now go to `client` directory and run the command `go run main.go`
6. The above would enable the client site
7. Now you can go to your browser and enter `http://localhost:8080/index.html` and access the service by logging in

### Note
The service is in no way complete and can be improved in many ways. This is still a work in progress so any changes required to code can be sent via pull requests.
