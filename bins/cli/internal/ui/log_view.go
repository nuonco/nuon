package ui

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go/models"
)

type LogState struct {
	State struct {
		Current string
	}
}

type LogTerminal struct {
	Terminal struct {
		Buffered bool
		Events   []struct {
			Line *struct {
				Msg string
			}
			Step *struct {
				Msg string
			}
		}
	}
}

func PrintBuildLog(log []models.ServiceBuildLog) {
	state := log[1]
	terminal := log[4]

	var trm LogTerminal
	err := mapstructure.Decode(terminal, &trm)
	if err != nil {
		PrintError(err)
		return
	}

	if len(trm.Terminal.Events) != 0 {
		for _, line := range trm.Terminal.Events {
			if line.Line != nil {
				if line.Line.Msg != "" {
					fmt.Println(line.Line.Msg)
				}
			}

			if line.Step != nil {
				if line.Step.Msg != "" {
					fmt.Println(line.Step.Msg)
				}
			}
		}
	} else {
		fmt.Println("Logs expire after 24hrs, run command again with --json to see full logs")
	}

	var ste LogState
	err = mapstructure.Decode(state, &ste)
	if err != nil {
		PrintError(err)
		return
	}

	fmt.Printf("status: %v\n", ste.State.Current)
}

func PrintDeployLogs(log []models.ServiceDeployLog) {
	state := log[1]
	terminal := log[4]

	var trm LogTerminal
	err := mapstructure.Decode(terminal, &trm)
	if err != nil {
		PrintError(err)
		return
	}

	if len(trm.Terminal.Events) != 0 {
		for _, line := range trm.Terminal.Events {
			if line.Line != nil {
				if line.Line.Msg != "" {
					fmt.Println(line.Line.Msg)
				}
			}

			if line.Step != nil {
				if line.Step.Msg != "" {
					fmt.Println(line.Step.Msg)
				}
			}
		}
	} else {
		fmt.Println("Logs expire after 24hrs, run command again with --json to see full logs")
	}

	var ste LogState
	err = mapstructure.Decode(state, &ste)
	if err != nil {
		PrintError(err)
		return
	}

	fmt.Printf("status: %v\n", ste.State.Current)
}
