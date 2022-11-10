package collect

//采集规则树
type RuleTree struct {
	Root  func() []*Request // 根节点(执行入口)
	Trunk map[string]*Rule  // 规则哈希表
}

// 采集规则节点
type Rule struct {
	ParseFunc func(*Context) ParseResult // 内容解析函数
}
