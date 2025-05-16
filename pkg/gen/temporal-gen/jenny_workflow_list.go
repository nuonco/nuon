package main

import (
	"bytes"
	"fmt"

	"github.com/grafana/codejen"
)

// WorkflowListJenny is a jenny that generates a function which returns function references to all the base functions underlying the Await wrappers
// generated in this run. This output is suitable for registration with a Temporal client.
type WorkflowListJenny struct {
	UseMethods bool
}

func (w WorkflowListJenny) JennyName() string {
	return "WorkflowListJenny"
}

func (w WorkflowListJenny) Generate(bfs ...*BaseFile) (*codejen.File, error) {
	// TODO(sdboyer) this is all skating by on assuming just one grouping of fns in the input set
	wfns := make([]WorkflowFn, 0)
	for _, bf := range bfs {
		wfns = append(wfns, bf.WorkflowFns...)
	}
	// Only generate lists when there are workflow fns that are methods
	if len(wfns) == 0 || wfns[0].Fn.Recv == nil {
		return nil, nil
	}

	buf := new(bytes.Buffer)
	genImports(buf, bfs[0])

	sym0 := extractFnSymbols(wfns[0].Fn)
	tvars := tvars_awaitfn_list{
		FnName:       "ListWorkflowFns",
		TemporalVerb: "Workflow",
		Receiver:     sym0.receiverTypeName,
		ReceiverLit:  sym0.receiverLit,
		ReceiverVar:  sym0.receiverVar,
	}

	if tvars.Receiver != "" {
		tvars.FnName += "On" + tvars.Receiver
	}

	for _, afn := range wfns {
		sym := extractFnSymbols(afn.Fn)
		if sym.receiverTypeName != tvars.Receiver {
			a, b := sym.receiverTypeName, tvars.Receiver
			if sym.receiverTypeName == "" {
				a = "standalone funcs"
			} else if tvars.Receiver == "" {
				b = "standalone funcs"
			}
			return nil, fmt.Errorf("generating fn list for more than group of activities at a time is currently unsupported (saw groups for %s and %s)", a, b)
		}
		// tvars.BaseFnNames = append(tvars.BaseFnNames, sym.fnSym)
		tvars.BaseFnNames = append(tvars.BaseFnNames, afn.Fn.Name.String())
	}

	err := tmpls.Lookup("workflow_list.tmpl").Execute(buf, tvars)
	if err != nil {
		return nil, fmt.Errorf("error executing activities list template: %w", err)
	}

	return codejen.NewFile("workflow_list_gen.go", buf.Bytes(), WorkflowListJenny{}), nil
}
