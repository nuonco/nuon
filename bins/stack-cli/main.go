// stack-cli is a thin wrapper around the installer SDK.
//
// All inputs come from the create-run URL the dashboard surfaces. The CLI
// POSTs to it, reads install_id / region / runner details from the response,
// and drives the SDK from there.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nuonco/nuon/sdks/stack"
)

// Version is set at build time via:
//
//	go build -ldflags "-X main.Version=v0.1.0" ./bins/stack-cli
//
// Local builds and `go run` report "dev".
var Version = "dev"

const usage = `stack-cli — provision/reprovision/deprovision a Nuon install stack.

Usage:
  stack-cli provision    <create-run-url>
  stack-cli reprovision  <create-run-url>
  stack-cli deprovision  <create-run-url>
  stack-cli status       --install-id <id> --region <region>
  stack-cli version      [--json]

The <create-run-url> is the POST endpoint the Nuon dashboard renders, e.g.
  https://api.nuon.co/v1/stack-runs/aws…
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "-h", "--help", "help":
		fmt.Print(usage)
	case "provision":
		runFromURL("provision")
	case "reprovision":
		runFromURL("reprovision")
	case "deprovision":
		runFromURL("deprovision")
	case "status":
		runStatus()
	case "version", "--version", "-v":
		runVersion()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}

func runFromURL(verb string) {
	if len(os.Args) < 3 {
		fail(fmt.Errorf("missing <create-run-url> for %s; see --help", verb))
	}
	url := os.Args[2]

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	inst, err := stack.FromURL(ctx, stack.URLOptions{URL: url, Kind: stack.Kind(verb)})
	if err != nil {
		fail(err)
	}

	var (
		opErr error
		st    any
	)
	switch verb {
	case "provision":
		st, opErr = inst.Provision(ctx)
	case "reprovision":
		st, opErr = inst.Reprovision(ctx)
	case "deprovision":
		opErr = inst.Deprovision(ctx)
	}

	// Flush logs with a fresh context — defer + os.Exit would skip this.
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	_ = inst.Close(closeCtx)
	closeCancel()

	if opErr != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %s\n", verb, opErr.Error())
		os.Exit(1)
	}
	if st != nil {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(st)
	}
}

func runStatus() {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	installID := fs.String("install-id", "", "install identifier (required)")
	region := fs.String("region", "us-east-1", "AWS region")
	_ = fs.Parse(os.Args[2:])
	if *installID == "" {
		fail(fmt.Errorf("--install-id is required"))
	}
	opts := stack.Options{InstallID: *installID, AWSRegion: *region}
	inst, err := stack.New(context.Background(), opts)
	if err != nil {
		fail(err)
	}
	defer inst.Close(context.Background())
	st, err := inst.Status()
	if err != nil {
		fail(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(st)
}

func runVersion() {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "emit JSON")
	_ = fs.Parse(os.Args[2:])
	if *asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]string{"version": Version})
		return
	}
	fmt.Println(Version)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
