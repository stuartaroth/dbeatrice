package main

type Connector interface {
	Name() string
	Query(input string) (*QueryResults, error)
	Execute(input string) (*ExecuteResult, error)
}

type QueryResults struct {
	Headers    []Header
	Rows       [][]string
	DurationMs int64
}

type Header struct {
	Name string
	Type string
}

type ExecuteResult struct {
	CountRowsAffected int
}
