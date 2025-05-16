package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"path/filepath"

	"github.com/go-toolsmith/astfmt"
	"github.com/grafana/codejen"
)

type tvars_asactivityfn struct {
	// FnName is the name of the generated method.
	FnName string

	// RequestTypeName is the name of the generated struct type used as the request parameter for the activity function
	RequestTypeName string

	// ActivitiesTypeName is the name of the generated struct type used as the receiver for all generated activity methods.
	ActivitiesTypeName string

	// BaseFnName is the name of the underlying function to be executed in an activity. It includes a receiver if the function is being generated as a method
	BaseFnName string

	// BaseFnSymbol is a string literal containing the name of the underlying activity function to be called. It includes a receiver if the function is being generated as a method
	// BaseFnSymbol string

	// RespType is a string literal representing the qualified type of the response
	RespType string

	// CallFuncName is the name of the generated function that will be generated (by another jenny) that the user should call to reach this one
	CallFuncName string

	// ParamFields is an ordered list of the field names in the request struct to be passed to the base function as direct arguments
	ParamFields []string
}

type AsActivityJenny struct {
	UseMethods bool
}

func (j AsActivityJenny) JennyName() string {
	return "AsActivityJenny"
}

func (w AsActivityJenny) Generate(bf *BaseFile) (*codejen.File, error) {
	if len(bf.AsActivityFns) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer
	genImports(&buf, bf)
	for _, asfn := range bf.AsActivityFns {
		var wv tvars_asactivityfn

		wv.BaseFnName = asfn.Fn.Name.Name

		gd, err := paramsToStruct(bf.Package.Fset, asfn.Fn)
		if err != nil {
			return nil, err
		}

		wv.RequestTypeName = gd.Specs[0].(*ast.TypeSpec).Name.Name

		wv.RespType = astfmt.Sprint(asfn.Fn.Type.Results.List[0].Type)

		fmt.Fprint(&buf, "\n"+astfmt.Sprint(gd), "\n")

		for _, param := range asfn.Fn.Type.Params.List[1:] {
			for _, name := range param.Names {
				wv.ParamFields = append(wv.ParamFields, titleize(name).Name)
			}
		}

		wv.FnName = asfn.Fn.Name.Name
		wv.ActivitiesTypeName = activityTypeFor(asfn.Fn)

		err = tmpls.Lookup("as_activity_fn.tmpl").Execute(&buf, wv)
		if err != nil {
			return nil, fmt.Errorf("error executing as_activity template for fn %s: %w", asfn.Fn.Name.String(), err)
		}

		// TODO CallFuncName
	}

	genpath := filepath.Base(bf.Path)
	genpath = genpath[:len(genpath)-3] + ".as_activity_gen.go"
	return codejen.NewFile(genpath, buf.Bytes(), AsActivityJenny{}), nil
}
