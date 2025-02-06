package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	cf "github.com/powertoolsdev/mono/pkg/gen/cfngen"
)

func main() {
	file := os.Args[1]
	byt, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	cfg := &cf.CfgFile{}
	err = toml.Unmarshal(byt, cfg)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	tmpl, err := cf.GenerateCloudformation(&cfg.VendorCfg, cfg.Internal)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	byt, err = tmpl.JSON()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(string(byt))
}
