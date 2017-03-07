package workflow

import "fmt"

type requestWorkflow struct {
	*wfBase
}

func newRequest(args Args) Workflow {
	return &requestWorkflow{
		wfBase: newBase(args),
	}
}

func (wf *requestWorkflow) Setup() error {
	fmt.Println("Request start!")
	return nil
}
