package spider

import "sync"

type ResourceRepository interface {
	Set(req map[string]*ResourceSpec)
	Add(req *ResourceSpec)
	Delete(name string)
	HasResource(name string) bool
}

type resourceRepository struct {
	resources map[string]*ResourceSpec
	rlock     sync.Mutex
}

func NewResourceRepository() ResourceRepository {
	r := &resourceRepository{}
	r.resources = make(map[string]*ResourceSpec, 100)
	return r
}

func (r *resourceRepository) Set(req map[string]*ResourceSpec) {
	r.rlock.Lock()
	defer r.rlock.Unlock()
	r.resources = req
}

func (r *resourceRepository) Add(req *ResourceSpec) {
	r.rlock.Lock()
	defer r.rlock.Unlock()
	r.resources[req.Name] = req
}

func (r *resourceRepository) Delete(name string) {
	r.rlock.Lock()
	defer r.rlock.Unlock()
	delete(r.resources, name)
}

func (r *resourceRepository) HasResource(name string) bool {
	r.rlock.Lock()
	defer r.rlock.Unlock()
	if _, ok := r.resources[name]; ok {
		return true
	}

	return false
}
