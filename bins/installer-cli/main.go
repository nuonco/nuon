// installer-cli is a thin wrapper around the installer SDK.
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

	"github.com/nuonco/nuon/sdks/nuon-installer-go/installer"
)

const usage = `installer-cli — provision/reprovision/deprovision a Nuon install stack.

Usage:
  installer-cli provision    <create-run-url>
  installer-cli reprovision  <create-run-url>
  installer-cli deprovision  <create-run-url>
  installer-cli status       --install-id <id> --region <region>

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

	inst, err := installer.FromCreateRunURL(ctx, installer.CreateRunURL{URL: url, Kind: verb})
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
	opts := installer.Options{InstallID: *installID, AWSRegion: *region}
	inst, err := installer.New(context.Background(), opts)
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

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
