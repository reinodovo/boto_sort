package bot

import (
	"log"

	"github.com/google/uuid"
	"github.com/reinodovo/boto-sort/internal/store"
)

type CompareRequest struct {
	id          string
	a, b        string
	chatId      int64
	response    chan int
	alreadySent bool
}

type CompareCallback func(req CompareRequest) error

type Comparator struct {
	requests        chan CompareRequest
	compareCallback CompareCallback
	pendingRequests map[string]CompareRequest
	store           *store.Store
}

func (c *Comparator) Compare(a, b string, chatId int64) int {
	resp := make(chan int)

	sorting, _ := c.store.GetSorting(chatId)
	for _, results := range sorting.CompareResults {
		if results.A == a && results.B == b {
			voteResult, unanimous := c.getVoteResult(sorting.Users, results.Votes)
			if unanimous {
				return voteResult
			}
			c.requests <- CompareRequest{id: results.Id, a: a, b: b, response: resp, chatId: chatId, alreadySent: true}
			return <-resp
		}
	}

	id := uuid.New().String()
	sorting.CompareResults[id] = store.CompareResult{
		Id:    id,
		A:     a,
		B:     b,
		Votes: make(map[int64]int),
	}
	c.store.SaveSorting(chatId, sorting)

	c.requests <- CompareRequest{id: id, a: a, b: b, response: resp, chatId: chatId, alreadySent: false}
	return <-resp
}

func NewComparator(compareCallback CompareCallback, store *store.Store) *Comparator {
	return &Comparator{requests: make(chan CompareRequest), compareCallback: compareCallback, store: store, pendingRequests: make(map[string]CompareRequest)}
}

func (c *Comparator) receiveVote(chatId int64, requestId string, userId int64, option string) error {
	result := -1
	if option == "b" {
		result = 1
	}

	sorting, err := c.store.GetSorting(chatId)
	if err != nil {
		return err
	}

	// ignore votes from users not participating in the poll
	if _, ok := sorting.Users[userId]; !ok {
		return nil
	}

	sorting.CompareResults[requestId].Votes[userId] = result
	err = c.store.SaveSorting(chatId, sorting)
	if err != nil {
		return err
	}

	votes := sorting.CompareResults[requestId].Votes
	result, unanimous := c.getVoteResult(sorting.Users, votes)
	if !unanimous {
		return nil
	}

	if req, ok := c.pendingRequests[requestId]; ok {
		req.response <- result
		delete(c.pendingRequests, requestId)
	}
	return nil
}

func (c *Comparator) getVoteResult(users map[int64]string, votes map[int64]int) (int, bool) {
	if len(users) != len(votes) {
		return 0, false
	}

	result := 1
	started := false
	for _, vote := range votes {
		if !started {
			result = vote
		}
		started = true
		if result != vote {
			return 0, false
		}
	}
	return result, true
}

func (c *Comparator) Start() {
	for req := range c.requests {
		c.pendingRequests[req.id] = req
		if req.alreadySent {
			continue
		}
		err := c.compareCallback(req)
		if err != nil {
			log.Println(err)
		}
	}
}
