package ui

import (
	"encoding/base64"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go/models"
)

type State struct {
	Current string
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Complete struct {
	Error Error `json:"error"`
}

type Terminal struct {
	Buffered bool
	Events   []struct {
		Line *struct {
			Msg string
		}
		Raw *struct {
			Data string
		}
		Step *struct {
			Msg    string
			Output string
		}
		Status *struct {
			Msg string
		}
	}
}

type Logs struct {
	Open     interface{}
	State    State
	Terminal Terminal
	Complete Complete
}

func PrintBuildLogs(logs []models.ServiceBuildLog) {
	var lgs Logs

	for _, l := range logs {
		err := mapstructure.Decode(l, &lgs)
		if err != nil {
			PrintError(err)
			return
		}

	}

	if len(lgs.Terminal.Events) != 0 {
		for _, line := range lgs.Terminal.Events {
			if line.Line != nil {
				if line.Line.Msg != "" {
					fmt.Println(line.Line.Msg)
				}
			}

			if line.Raw != nil {
				if line.Raw.Data != "" {
					l, err := base64.StdEncoding.DecodeString(line.Raw.Data)
					if err != nil {
						PrintError(err)
						return
					}
					fmt.Printf("%s\n", l)
				}
			}

			if line.Step != nil {
				if line.Step.Msg != "" {
					fmt.Println(line.Step.Msg)
				}

				if line.Step.Output != "" {
					data, err := base64.StdEncoding.DecodeString(line.Step.Output)
					if err != nil {
						PrintError(err)
						continue
					}
					fmt.Println(string(data))
				}
			}

			if line.Status != nil {
				if line.Status.Msg != "" {
					fmt.Println(line.Status.Msg)
				}
			}
		}
	} else {
		fmt.Println("Logs expire after 24hrs, run command again with --json to see full logs")
	}
	if lgs.Complete.Error != (Error{}) {
		fmt.Printf("job-error: %d\n", lgs.Complete.Error.Code)
		fmt.Printf("message: %s\n", lgs.Complete.Error.Message)
	}

	fmt.Printf("\nstatus: %v\n", lgs.State.Current)
}

func PrintDeployLogs(log []models.ServiceDeployLog) {
	var lgs Logs

	for _, l := range log {
		err := mapstructure.Decode(l, &lgs)
		if err != nil {
			PrintError(err)
			return
		}

	}

	if len(lgs.Terminal.Events) != 0 {
		for _, line := range lgs.Terminal.Events {
			if line.Line != nil {
				if line.Line.Msg != "" {
					fmt.Println(line.Line.Msg)
				}
			}

			if line.Raw != nil {
				if line.Raw.Data != "" {
					l, err := base64.StdEncoding.DecodeString(line.Raw.Data)
					if err != nil {
						PrintError(err)
						return
					}
					fmt.Printf("%s\n", l)
				}
			}

			if line.Step != nil {
				if line.Step.Msg != "" {
					fmt.Println(line.Step.Msg)
				}

				if line.Step.Output != "" {
					data, err := base64.StdEncoding.DecodeString(line.Step.Output)
					if err != nil {
						PrintError(err)
						continue
					}
					fmt.Println(string(data))
				}
			}

			if line.Status != nil {
				if line.Status.Msg != "" {
					fmt.Println(line.Status.Msg)
				}
			}
		}
	} else {
		fmt.Println("Logs expire after 24hrs, run command again with --json to see full logs")
	}
	if lgs.Complete.Error != (Error{}) {
		fmt.Printf("job-error: %d\n", lgs.Complete.Error.Code)
		fmt.Printf("message: %s\n", lgs.Complete.Error.Message)
	}

	fmt.Printf("\nstatus: %v\n", lgs.State.Current)
}

func PrintLogsFromInterface(log []interface{}) {
	var lgs Logs

	for _, l := range log {
		err := mapstructure.Decode(l, &lgs)
		if err != nil {
			PrintError(err)
			return
		}

	}

	if len(lgs.Terminal.Events) != 0 {
		for _, line := range lgs.Terminal.Events {
			if line.Line != nil {
				if line.Line.Msg != "" {
					fmt.Println(line.Line.Msg)
				}
			}

			if line.Raw != nil {
				if line.Raw.Data != "" {
					l, err := base64.StdEncoding.DecodeString(line.Raw.Data)
					if err != nil {
						PrintError(err)
						return
					}
					fmt.Printf("%s\n", l)
				}
			}

			if line.Step != nil {
				if line.Step.Msg != "" {
					fmt.Println(line.Step.Msg)
				}

				if line.Step.Output != "" {
					data, err := base64.StdEncoding.DecodeString(line.Step.Output)
					if err != nil {
						PrintError(err)
						continue
					}
					fmt.Println(string(data))
				}
			}

			if line.Status != nil {
				if line.Status.Msg != "" {
					fmt.Println(line.Status.Msg)
				}
			}
		}
	} else {
		fmt.Println("Logs expire after 24hrs, run command again with --json to see full logs")
	}
	if lgs.Complete.Error != (Error{}) {
		fmt.Printf("job-error: %d\n", lgs.Complete.Error.Code)
		fmt.Printf("message: %s\n", lgs.Complete.Error.Message)
	}

	fmt.Printf("\nstatus: %v\n", lgs.State.Current)
}
