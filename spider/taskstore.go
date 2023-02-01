package spider

import (
	"github.com/robertkrimen/otto"
)

// TaskStore is a global instace
var (
	TaskStore = &taskStore{
		List: []*Task{},
		Hash: map[string]*Task{},
	}
)

type taskStore struct {
	List []*Task
	Hash map[string]*Task
}

func (c *taskStore) Add(task *Task) {
	c.Hash[task.Name] = task
	c.List = append(c.List, task)
}

func (c *taskStore) AddJSTask(m *TaskModle) {
	task := &Task{
		//Property: m.Property,
	}

	task.Rule.Root = func() ([]*Request, error) {
		vm := otto.New()
		if err := vm.Set("AddJsReq", AddJsReqs); err != nil {
			return nil, err
		}

		v, err := vm.Eval(m.Root)

		if err != nil {
			return nil, err
		}

		e, err := v.Export()

		if err != nil {
			return nil, err
		}

		return e.([]*Request), nil
	}

	for _, r := range m.Rules {
		paesrFunc := func(parse string) func(ctx *Context) (ParseResult, error) {
			return func(ctx *Context) (ParseResult, error) {
				vm := otto.New()
				if err := vm.Set("ctx", ctx); err != nil {
					return ParseResult{}, err
				}

				v, err := vm.Eval(parse)

				if err != nil {
					return ParseResult{}, err
				}

				e, err := v.Export()

				if err != nil {
					return ParseResult{}, err
				}

				if e == nil {
					return ParseResult{}, err
				}

				return e.(ParseResult), err
			}
		}(r.ParseFunc)

		if task.Rule.Trunk == nil {
			task.Rule.Trunk = make(map[string]*Rule, 0)
		}

		task.Rule.Trunk[r.Name] = &Rule{
			ParseFunc: paesrFunc,
		}
	}

	c.Hash[task.Name] = task
	c.List = append(c.List, task)
}

// 用于动态规则添加多个请求。
func AddJsReqs(jreqs []map[string]interface{}) []*Request {
	reqs := make([]*Request, 0)

	for _, jreq := range jreqs {
		req := &Request{}
		u, ok := jreq["URL"].(string)

		if !ok {
			return nil
		}

		req.URL = u
		req.RuleName, _ = jreq["RuleName"].(string)
		req.Method, _ = jreq["Method"].(string)
		req.Priority, _ = jreq["Priority"].(int64)
		reqs = append(reqs, req)
	}

	return reqs
}

// 用于动态规则添加单个请求。
func AddJsReq(jreq map[string]interface{}) []*Request {
	reqs := make([]*Request, 0)
	req := &Request{}
	u, ok := jreq["URL"].(string)

	if !ok {
		return nil
	}

	req.URL = u
	req.RuleName, _ = jreq["RuleName"].(string)
	req.Method, _ = jreq["Method"].(string)
	req.Priority, _ = jreq["Priority"].(int64)
	reqs = append(reqs, req)

	return reqs
}

func GetFields(taskName string, ruleName string) []string {
	return TaskStore.Hash[taskName].Rule.Trunk[ruleName].ItemFields
}
