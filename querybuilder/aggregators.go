package querybuilder

import "fmt"

type functionCall func(string) string

type aggregators struct {
	Min   functionCall
	Max   functionCall
	Count functionCall
	Avg   functionCall
	Sum   functionCall
}

var Aggregators = &aggregators{
	Min:   makeFunctionFormatter("MIN"),
	Max:   makeFunctionFormatter("MAX"),
	Count: makeFunctionFormatter("COUNT"),
	Avg:   makeFunctionFormatter("AVG"),
	Sum:   makeFunctionFormatter("SUM"),
}

func makeFunctionFormatter(function string) func(string) string {
	return func(column string) string {
		return fmt.Sprintf("%s(%s)", function, column)
	}
}
