package spider

// 采集规则树
type RuleTree struct {
	Root  func() ([]*Request, error) // 根节点(执行入口)
	Trunk map[string]*Rule           // 规则哈希表
}

// 采集规则节点
type Rule struct {
	ItemFields []string
	ParseFunc  func(*Context) (ParseResult, error) // 内容解析函数
}

type ParseResult struct {
	Requesrts []*Request
	Items     []interface{}
}

// parse javascript
type (
	TaskModle struct {
		Property
		Root  string      `json:"root_script"`
		Rules []RuleModle `json:"rule"`
	}
	RuleModle struct {
		Name      string `json:"name"`
		ParseFunc string `json:"parse_script"`
	}
)

type Property struct {
	Name     string `json:"name"` // 任务名称，应保证唯一性
	URL      string `json:"url"`
	Cookie   string `json:"cookie"`
	WaitTime int64  `json:"wait_time"` // 随机休眠时间，秒
	Reload   bool   `json:"reload"`    // 网站是否可以重复爬取
	MaxDepth int64  `json:"max_depth"`
}
