// installer-cli is a thin wrapper around the installer SDK for manual testing.
//
// All inputs come from explicit flags — the dashboard renders a copy-paste
// command with values pre-filled, the same way it shows the CFN quick-link
// URL today. No API token, no config file.
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

const usage = `installer-cli — provision/deprovision a Nuon install stack via the AWS Go SDK.

Usage:
  installer-cli <command> [flags]

Commands:
  provision     Create the install stack in AWS.
  deprovision   Tear down the install stack.
  status        Print the persisted state file for an install.

Flags (all required for provision/deprovision unless noted):
  --install-id      Install identifier.
  --phone-home-id   Per-stack-version secret used as the URL path token.
  --region          AWS region.
  --ctl-api-url     ctl-api base URL (default https://api.nuon.co).
  --stdout-only     Skip ctl-api run reporting; log to stdout only.
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
		runProvision()
	case "deprovision":
		runDeprovision()
	case "status":
		runStatus()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}

type commonFlags struct {
	installID   string
	phoneHomeID string
	region      string
	ctlAPIURL   string
	stdoutOnly  bool
}

func bindCommon(fs *flag.FlagSet, c *commonFlags) {
	fs.StringVar(&c.installID, "install-id", "", "install identifier (required)")
	fs.StringVar(&c.phoneHomeID, "phone-home-id", "", "stack version phone-home ID (required unless --stdout-only)")
	fs.StringVar(&c.region, "region", "", "AWS region (required)")
	fs.StringVar(&c.ctlAPIURL, "ctl-api-url", "https://api.nuon.co", "ctl-api base URL")
	fs.BoolVar(&c.stdoutOnly, "stdout-only", false, "skip run reporting; log locally")
}

func (c *commonFlags) toOptions() (installer.Options, error) {
	if c.installID == "" {
		return installer.Options{}, fmt.Errorf("--install-id is required")
	}
	if c.region == "" {
		return installer.Options{}, fmt.Errorf("--region is required")
	}
	o := installer.Options{InstallID: c.installID, AWSRegion: c.region}
	if !c.stdoutOnly {
		if c.phoneHomeID == "" {
			return installer.Options{}, fmt.Errorf("--phone-home-id is required (or pass --stdout-only)")
		}
		o.StackRun = &installer.StackRunConfig{
			CtlAPIURL:   c.ctlAPIURL,
			PhoneHomeID: c.phoneHomeID,
		}
	}
	return o, nil
}

func runProvision() {
	fs := flag.NewFlagSet("provision", flag.ExitOnError)
	c := &commonFlags{}
	bindCommon(fs, c)
	_ = fs.Parse(os.Args[2:])

	opts, err := c.toOptions()
	if err != nil {
		fail(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	inst, err := installer.New(ctx, opts)
	if err != nil {
		fail(err)
	}

	st, provErr := inst.Provision(ctx)
	// Flush logs with a fresh, uncanceled context — defer + os.Exit would skip
	// this and the OTLP batch processor would never push.
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	_ = inst.Close(closeCtx)
	closeCancel()

	if provErr != nil {
		fmt.Fprintln(os.Stderr, "provision failed:", provErr)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(st)
}

func runDeprovision() {
	fs := flag.NewFlagSet("deprovision", flag.ExitOnError)
	c := &commonFlags{}
	bindCommon(fs, c)
	_ = fs.Parse(os.Args[2:])

	opts, err := c.toOptions()
	if err != nil {
		fail(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	inst, err := installer.New(ctx, opts)
	if err != nil {
		fail(err)
	}

	deprovErr := inst.Deprovision(ctx)
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	_ = inst.Close(closeCtx)
	closeCancel()

	if deprovErr != nil {
		fmt.Fprintln(os.Stderr, "deprovision failed:", deprovErr)
		os.Exit(1)
	}
}

func runStatus() {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	installID := fs.String("install-id", "", "install identifier (required)")
	_ = fs.Parse(os.Args[2:])
	if *installID == "" {
		fail(fmt.Errorf("--install-id is required"))
	}
	opts := installer.Options{InstallID: *installID, AWSRegion: "us-east-1"}
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
