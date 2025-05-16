package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/grafana/codejen"
)

// TODO(sdboyer) the activity and workflow list jennies are diverging, and this template var should be split out accordingly
type tvars_awaitfn_list struct {
	// FnName is the name of the generated list function
	FnName string

	// Receiver is the var declaration and name of the receiver type. Populated only if the underlying base functions were declared as methods on a type.
	Receiver string
	// ReceiverLit is the var declaration and name of the receiver type. Populated only if the underlying base functions were declared as methods on a type.
	ReceiverLit string
	// ReceiverVar is the name of the var used for the receiver, for use in function scope. Populated only if the underlying base functions were declared as methods on a type.
	ReceiverVar string

	// BaseFnNames is a list of the underlying function names to be returned.
	BaseFnNames []string

	// TemporalVerb is the type of temporal action, "Workflow" or "Activity". Used for docs.
	TemporalVerb string
}

// ActivityListJenny is a jenny that generates a function which returns function references to all the base functions underlying the Await wrappers
// generated in this run. This output is suitable for registration with a Temporal client.
type ActivityListJenny struct {}

func (w ActivityListJenny) JennyName() string {
	return "ActivityListJenny"
}

func (w ActivityListJenny) Generate(bfs ...*BaseFile) (*codejen.File, error) {
	afns := make([]ActivityFn, 0)
	for _, bf := range bfs {
		afns = append(afns, bf.ActivityFns...)
	}
	if len(afns) == 0 {
		return nil, nil
	}
	buf := new(bytes.Buffer)
	genImports(buf, bfs[0])

	tvars := tvars_awaitfn_list{
		FnName:       "ActivityFns",
		TemporalVerb: "Activity",
		Receiver:     extractFnSymbols(afns[0].Fn).receiverTypeName,
	}

	if tvars.Receiver != "" {
		tvars.FnName = strings.TrimPrefix(tvars.Receiver+"ActivityFns", "*")
	}

	for _, afn := range afns {
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

	err := tmpls.Lookup("activity_list.tmpl").Execute(buf, tvars)
	if err != nil {
		return nil, fmt.Errorf("error executing activities list template: %w", err)
	}

	return codejen.NewFile("activity_list_gen.go", buf.Bytes(), ActivityListJenny{}), nil
}
