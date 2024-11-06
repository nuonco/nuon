package wrt

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/fs"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"go.uber.org/zap"
)

type ModuleConfig struct {
	wazero.ModuleConfig
	Logger *zap.Logger
	Fsys   fs.FS
	// ModuleName is the name of the module that will be compiled and executed
	ModuleName string
	Env        map[string]string
}

// NewRuntime returns a new Runtime instance with the standard Nuon modifications made to the wazero runtime.
func NewRuntime(ctx context.Context) wazero.Runtime {
	cfg := wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true)
	rt := wazero.NewRuntimeWithConfig(ctx, cfg)

	wasi_snapshot_preview1.MustInstantiate(ctx, rt)
	return rt
}

// ExecModuleWithDefaults compiles, instantiates, and runs the entry function of a WebAssembly module using Nuon's standard configuration.
func ExecModuleWithDefaults(ctx context.Context, rt wazero.Runtime, wasmbytes []byte, cfg ModuleConfig) (api.Module, error) {
	guest, err := rt.CompileModule(ctx, wasmbytes)
	if err != nil {
		panic(fmt.Sprintf("TODO: errs, %s", err))
	}
	// for n, f := range guest.ExportedFunctions() {
	// 	fmt.Println(n, f.ExportNames(), f.ParamNames(), f.ParamTypes(), f.DebugName())
	// }

	mcfg := wazero.NewModuleConfig().
		WithRandSource(rand.Reader).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithArgs(cfg.ModuleName)

	if cfg.Logger != nil {
		mcfg = mcfg.WithStdout(&logger{name: cfg.ModuleName, zl: cfg.Logger}).
			WithStderr(&logger{name: cfg.ModuleName, zl: cfg.Logger, isStderr: true})
	}

	if cfg.Fsys != nil {
		mcfg = mcfg.WithFS(cfg.Fsys)
	}
	for k, v := range cfg.Env {
		mcfg = mcfg.WithEnv(k, v)
	}

	var mod api.Module
	switch detectImports(guest.ImportedFunctions()) {
	case modeWasiUnstable:
		panic("TODO: errs; let's not support wasi_unstable?")
	case modeWasi, modeDefault:
		// We assume wasi was already instantiated
		mod, err = rt.InstantiateModule(ctx, guest, mcfg)
	}

	if err != nil {
		panic(fmt.Sprintf("TODO: errs, %s", err))
	}

	// Instantiating the module will run the _start function (this is wazero's default, overridable with mcfg.StartFunctions()).
	// As long as running the start function is all we want to do, we're done.
	// TODO(sdboyer) - figure out if _start is a general standard in wasm-world or if we're staking out a position by following wazero
	return mod, nil
}

func detectImports(imports []api.FunctionDefinition) importMode {
	for _, f := range imports {
		moduleName, _, _ := f.Import()
		switch moduleName {
		case wasi_snapshot_preview1.ModuleName:
			return modeWasi
		case "wasi_unstable":
			return modeWasiUnstable
		}
	}
	return modeDefault
}

const (
	modeDefault importMode = iota
	modeWasi
	modeWasiUnstable
)

type importMode uint
