package workflow

import "fmt"

type requestWorkflow struct {
	*wfBase
}

func newRequest(host Host) Workflow {
	return &requestWorkflow{
		wfBase: newBase(host),
	}
}

func (wf *requestWorkflow) Setup() error {
	fmt.Println("Request start!")
	return nil
}

func (wf *requestWorkflow) Listen() {

}
