package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/file"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// Param is defined in activity_generator.go

// GeneratorOptions contains configuration for the code generator
type GeneratorOptions struct {
	ProcessImports bool
}

func GenerateForFile(f *file.File, opts GeneratorOptions) error {
	var body bytes.Buffer
	hasActivity := false
	hasWorkflow := false
	hasTime := false
	hasTemplate := false
	hasActivityWrapper := false
	hasClient := false
	hasNamespacedActivity := false

	hasTemporal := false
	namespacedActivities := []ActivityData{}

	filename := filepath.Base(f.Path)
	fileExt := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, fileExt)
	clientName := toPascalCase(baseName)

	for _, fn := range f.Functions {
		inputType, outputType, params, receiver, err := getSignature(f.Package.Pkg, fn.Decl)
		if err != nil {
			return fmt.Errorf("failed to get signature for %s: %w", fn.Decl.Name.Name, err)
		}

		var code []byte
		if fn.Annotation.Type == "activity" {
			hasActivity = true
			if fn.Annotation.ActivityOpts.GenerateWrapper {
				hasActivityWrapper = true
			}
			if fn.Annotation.ActivityOpts.ScheduleToCloseTimeout > 0 || fn.Annotation.ActivityOpts.StartToCloseTimeout > 0 {
				hasTime = true
			}
			if fn.Annotation.ActivityOpts.MaxRetries > 0 {
				hasTemporal = true
			}

			var byFieldType string
			if fn.Annotation.ActivityOpts.ByField != "" {
				if fn.Annotation.ActivityOpts.GenerateWrapper {
					// Look in params for the field
					fieldName := fn.Annotation.ActivityOpts.ByField
					found := false
					for _, p := range params {
						if p.ExportedName == fieldName || p.Name == fieldName {
							byFieldType = p.Type
							// Update ByField to use the exported name for the struct field assignment
							fn.Annotation.ActivityOpts.ByField = p.ExportedName
							found = true
							break
						}
					}
					if !found {
						return fmt.Errorf("field %s not found in generated wrapper parameters", fieldName)
					}
				} else {
					obj := f.Package.Pkg.TypesInfo.Defs[fn.Decl.Name]
					if sig, ok := obj.Type().(*types.Signature); ok {
						sigParams := sig.Params()
						start := 0
						if sigParams.Len() > 0 {
							firstParamType := sigParams.At(0).Type().String()
							if strings.Contains(firstParamType, "context.Context") || strings.Contains(firstParamType, "workflow.Context") {
								start = 1
							}
						}
						if start < sigParams.Len() {
							inputParam := sigParams.At(start)
							fieldType, err := getFieldType(inputParam.Type(), fn.Annotation.ActivityOpts.ByField)
							if err != nil {
								return fmt.Errorf("failed to find field %s in %s: %w", fn.Annotation.ActivityOpts.ByField, inputParam.Type(), err)
							}

							qualifier := func(p *types.Package) string {
								if p == f.Package.Pkg.Types {
									return ""
								}
								return p.Name()
							}
							byFieldType = types.TypeString(fieldType, qualifier)
						}
					}
				}
			}

			qualifiedName := fn.Decl.Name.Name
			if fn.Annotation.ActivityOpts.Namespace != "" {
				qualifiedName = fn.Annotation.ActivityOpts.Namespace + "." + fn.Decl.Name.Name
			}

			data := ActivityData{
				Name:          fn.Decl.Name.Name,
				OriginalName:  fn.Decl.Name.Name,
				QualifiedName: qualifiedName,
				InputType:     inputType,
				OutputType:    outputType,
				Options:       fn.Annotation.ActivityOpts,
				Params:        params,
				Receiver:      receiver,
				ByFieldType:   byFieldType,
			}
			code, err = GenerateActivity(data)

			// Track namespaced activities for registration generation
			if fn.Annotation.ActivityOpts.Namespace != "" {
				hasNamespacedActivity = true
				namespacedActivities = append(namespacedActivities, data)
			}
		} else if fn.Annotation.Type == "workflow" {
			hasWorkflow = true
			if fn.Annotation.WorkflowOpts.ExecutionTimeout > 0 || fn.Annotation.WorkflowOpts.TaskTimeout > 0 {
				hasTime = true
			}
			if fn.Annotation.WorkflowOpts.IDTemplate != "" {
				hasTemplate = true
			}
			data := WorkflowData{
				Name:         fn.Decl.Name.Name,
				OriginalName: fn.Decl.Name.Name,
				InputType:    inputType,
				OutputType:   outputType,
				Options:      fn.Annotation.WorkflowOpts,
				Receiver:     receiver,
			}
			code, err = GenerateWorkflow(data)
		} else if fn.Annotation.Type == "query" {
			hasClient = true
			data := QueryData{
				ClientName:   clientName,
				Name:         fn.Decl.Name.Name,
				OriginalName: fn.Decl.Name.Name,
				InputType:    inputType,
				OutputType:   outputType,
				Options:      fn.Annotation.QueryOpts,
			}
			code, err = GenerateQuery(data)
		} else if fn.Annotation.Type == "update" {
			hasClient = true
			updateName := fn.Decl.Name.Name
			if fn.Annotation.UpdateOpts != nil && fn.Annotation.UpdateOpts.ID != "" {
				updateName = fn.Annotation.UpdateOpts.ID
			}
			data := UpdateData{
				ClientName:   clientName,
				Name:         fn.Decl.Name.Name,
				OriginalName: fn.Decl.Name.Name,
				UpdateName:   updateName,
				InputType:    inputType,
				OutputType:   outputType,
				Options:      fn.Annotation.UpdateOpts,
			}
			code, err = GenerateUpdate(data)
		} else if fn.Annotation.Type == "signal" {
			hasClient = true
			data := SignalData{
				ClientName:   clientName,
				Name:         fn.Decl.Name.Name,
				OriginalName: fn.Decl.Name.Name,
				InputType:    inputType,
				Options:      fn.Annotation.SignalOpts,
			}
			code, err = GenerateSignal(data)
		}

		if err != nil {
			return fmt.Errorf("failed to generate code for %s: %w", fn.Decl.Name.Name, err)
		}

		if code != nil {
			body.Write(code)
			body.WriteString("\n")
		}
	}

	if body.Len() == 0 {
		return nil
	}

	var out bytes.Buffer
	// Header
	out.WriteString("//  THIS FILE IS GENERATED. DO NOT EDIT.\n")
	out.WriteString(fmt.Sprintf("//  %s\n\n", config.Watermark))
	out.WriteString(fmt.Sprintf("package %s\n\n", f.Package.Pkg.Name))
	out.WriteString("import (\n")
	if hasTemplate {
		out.WriteString("\t\"bytes\"\n")
		out.WriteString("\t\"fmt\"\n")
		out.WriteString("\t\"text/template\"\n")
	}
	if hasActivityWrapper || hasClient {
		out.WriteString("\t\"context\"\n")
	}
	if hasTime {
		out.WriteString("\t\"time\"\n")
	}
	out.WriteString("\n")
	if hasClient {
		out.WriteString("\t\"go.temporal.io/sdk/client\"\n")
	}
	if hasActivity || hasWorkflow {
		out.WriteString("\t\"go.temporal.io/sdk/workflow\"\n")
	}
	if hasTemporal {
		out.WriteString("\t\"go.temporal.io/sdk/temporal\"\n")
	}
	if hasNamespacedActivity {
		out.WriteString("\t\"go.temporal.io/sdk/activity\"\n")
		out.WriteString("\t\"go.temporal.io/sdk/worker\"\n")
	}
	out.WriteString(")\n\n")

	// If we have a client, generate the client struct and options
	if hasClient {
		clientCode, err := GenerateClient(ClientData{
			ClientName: clientName,
		})
		if err != nil {
			return fmt.Errorf("failed to generate client code: %w", err)
		}
		out.Write(clientCode)
		out.WriteString("\n")
	}

	out.Write(body.Bytes())

	// Generate registration helpers for namespaced activities
	if hasNamespacedActivity {
		for _, actData := range namespacedActivities {
			regCode, err := GenerateActivityRegistration(actData)
			if err != nil {
				return fmt.Errorf("failed to generate registration for %s: %w", actData.Name, err)
			}
			out.Write(regCode)
			out.WriteString("\n")
		}
	}

	// Write to file
	ext := filepath.Ext(f.Path)
	base := strings.TrimSuffix(f.Path, ext)
	outPath := base + "_gen.go"

	// Conditionally process imports if flag is enabled
	var finalBytes []byte
	if opts.ProcessImports {
		// Parse original to get imports before processing
		fset := token.NewFileSet()
		original, err := parser.ParseFile(fset, outPath, out.Bytes(), parser.ParseComments)
		if err != nil {
			return fmt.Errorf("error parsing generated file: %w", err)
		}

		// Process imports
		formatted, err := imports.Process(outPath, out.Bytes(), nil)
		if err != nil {
			return fmt.Errorf("goimports processing failed for %s: %w", outPath, err)
		}

		// Parse result to check for added imports
		processed, err := parser.ParseFile(fset, outPath, formatted, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("error parsing formatted file: %w", err)
		}

		// Build map of original imports
		originalImports := make(map[string]bool)
		for _, imp := range original.Imports {
			originalImports[imp.Path.Value] = true
		}

		// Check for added imports
		var addedImports []string
		for _, imp := range processed.Imports {
			if !originalImports[imp.Path.Value] {
				addedImports = append(addedImports, imp.Path.Value)
			}
		}

		// Allow goimports to add imports for domain types (e.g., app, signal)
		// that are referenced in generated code but not in the hardcoded import list.
		_ = addedImports

		finalBytes = formatted
	} else {
		finalBytes = out.Bytes()
	}

	if err := os.WriteFile(outPath, finalBytes, 0644); err != nil {
		return err
	}

	return nil
}

// Package needs to be imported from internal/dir but types.Package is used here
// We need to pass the types.Package directly or wrapper
// Actually f.Package is *dir.Package which contains Pkg *packages.Package.
// packages.Package contains Types *types.Package.

func getSignature(pkg *packages.Package, decl *ast.FuncDecl) (inputType string, outputType string, params []Param, receiver string, err error) {
	obj := pkg.TypesInfo.Defs[decl.Name]
	if obj == nil {
		return "", "", nil, "", fmt.Errorf("type object not found for %s", decl.Name.Name)
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return "", "", nil, "", fmt.Errorf("not a function")
	}

	qualifier := func(p *types.Package) string {
		if p == pkg.Types {
			return ""
		}
		return p.Name()
	}

	// Receiver
	if recv := sig.Recv(); recv != nil {
		receiver = types.TypeString(recv.Type(), qualifier)
	}

	// Inputs
	sigParams := sig.Params()
	start := 0
	// Skip context (first arg)
	if sigParams.Len() > 0 {
		// Check if first arg is context.Context or workflow.Context
		firstParamType := sigParams.At(0).Type().String()
		if strings.Contains(firstParamType, "context.Context") || strings.Contains(firstParamType, "workflow.Context") {
			start = 1
		}
	}

	for i := start; i < sigParams.Len(); i++ {
		param := sigParams.At(i)
		typeName := types.TypeString(param.Type(), qualifier)
		name := param.Name()
		params = append(params, Param{
			Name:         name,
			Type:         typeName,
			ExportedName: toPascal(name),
		})
	}

	if len(params) > 0 {
		inputType = params[0].Type
	}

	// Outputs
	results := sig.Results()
	if results.Len() > 0 {
		// Assume last is error.
		if results.Len() == 2 {
			outputType = types.TypeString(results.At(0).Type(), qualifier)
		} else if results.Len() == 1 {
			// Just error
			outputType = ""
		}
	}

	return inputType, outputType, params, receiver, nil
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func toPascal(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func getFieldType(t types.Type, fieldName string) (types.Type, error) {
	// Dereference pointer if needed
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// Check if it's a named type
	named, ok := t.(*types.Named)
	if !ok {
		return nil, fmt.Errorf("type %s is not a named type", t)
	}

	// Check if underlying is a struct
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("type %s is not a struct", t)
	}

	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		if field.Name() == fieldName {
			return field.Type(), nil
		}
	}

	return nil, fmt.Errorf("field %s not found in %s", fieldName, t)
}
