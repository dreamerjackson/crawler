package collect

type Request struct {
	Url       string
	Cookie    string
	ParseFunc func([]byte, *Request) ParseResult
}

type ParseResult struct {
	Requesrts []*Request
	Items     []interface{}
}
