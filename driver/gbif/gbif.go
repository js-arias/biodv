// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package gbif implements an interface
// to GBIF webservice.
package gbif

import (
	"bytes"
	"net/http"
	"time"
)

// Retry is the number of times a request will be retried
// before aborted.
var Retry = 5

// Timeout is the timeout of the http request.
var Timeout = 20 * time.Second

// Wait is the waiting time for a new request
// (we don't want to overload the GBIF server!).
var Wait = time.Millisecond * 300

// Buffer is the maximum number of requests in the request queue.
var Buffer = 100

const wsHead = "http://api.gbif.org/v1/"

// Request contains an gbif request,
// and a channel with the answers.
type request struct {
	req string
	ans chan bytes.Buffer
	err chan error
}

// NewRequest sends a request to the request channel.
func newRequest(req string) request {
	r := request{
		req: wsHead + req,
		ans: make(chan bytes.Buffer),
		err: make(chan error),
	}
	reqChan.cReqs <- r
	return r
}

// ReqChanType keeps the requests channel.
type reqChanType struct {
	cReqs chan request
}

// ReqChan is the requests channel.
// It should be initialized before
// using the database.
var reqChan *reqChanType

// InitReqs initialize the request channel.
func initReqs() {
	http.DefaultClient.Timeout = Timeout
	reqChan = &reqChanType{cReqs: make(chan request, Buffer)}
	go reqChan.reqs()
}

// Reqs make the network request.
func (rc *reqChanType) reqs() {
	for r := range rc.cReqs {
		a, err := http.Get(r.req)
		if err != nil {
			r.err <- err
			continue
		}
		var b bytes.Buffer
		b.ReadFrom(a.Body)
		a.Body.Close()
		r.ans <- b

		// we do not want to overload the gbif server.
		time.Sleep(Wait)
	}
}

// Database is handler of the GBIF database.
type database struct{}
