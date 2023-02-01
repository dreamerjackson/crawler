package spider

import "sync"

type ReqHistoryRepository interface {
	AddVisited(reqs ...*Request)
	DeleteVisited(req *Request)
	AddFailures(req *Request) bool
	DeleteFailures(req *Request)
	HasVisited(req *Request) bool
}

type reqHistory struct {
	Visited     map[string]bool
	VisitedLock sync.Mutex

	failures    map[string]*Request // 失败请求id -> 失败请求
	failureLock sync.Mutex
}

func NewReqHistoryRepository() ReqHistoryRepository {
	r := &reqHistory{}
	r.Visited = make(map[string]bool, 100)
	r.failures = make(map[string]*Request, 100)
	return r
}

func (r *reqHistory) HasVisited(req *Request) bool {
	r.VisitedLock.Lock()
	defer r.VisitedLock.Unlock()

	unique := req.Unique()

	return r.Visited[unique]
}

func (r *reqHistory) AddVisited(reqs ...*Request) {
	r.VisitedLock.Lock()
	defer r.VisitedLock.Unlock()

	for _, req := range reqs {
		unique := req.Unique()
		r.Visited[unique] = true
	}
}

func (r *reqHistory) DeleteVisited(req *Request) {
	r.VisitedLock.Lock()
	defer r.VisitedLock.Unlock()
	unique := req.Unique()
	delete(r.Visited, unique)
}

func (r *reqHistory) AddFailures(req *Request) bool {

	first := true
	if !req.Task.Reload {
		r.DeleteVisited(req)
	}

	r.failureLock.Lock()
	defer r.failureLock.Unlock()

	if _, ok := r.failures[req.Unique()]; !ok {
		r.failures[req.Unique()] = req
	} else {
		first = false
	}

	return first

}

func (r *reqHistory) DeleteFailures(req *Request) {
	r.failureLock.Lock()
	defer r.failureLock.Unlock()

	delete(r.failures, req.Unique())
}
