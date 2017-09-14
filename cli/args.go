package cli

import (
	"fmt"
	"strings"
)

type DiagnosticFunc func(message string, command string)

type ArgConsumer struct {
	positionalArgs []string
	command        string
	consumed       map[int]struct{}
	diagnose       DiagnosticFunc
}

func NewArgConsumer(positionalArgs []string, diagnose DiagnosticFunc) *ArgConsumer {
	consumed := make(map[int]struct{})
	consumed[0] = struct{}{}
	return &ArgConsumer{
		positionalArgs: positionalArgs,
		command:        positionalArgs[0],
		consumed:       consumed,
		diagnose:       diagnose,
	}
}

func (ac *ArgConsumer) Consume(arg int, argDescription string) string {
	if len(ac.positionalArgs) < arg+1 || ac.positionalArgs[arg] == "" {
		ac.diagnose(fmt.Sprintf("Incorrect usage: %s not specified.", argDescription), ac.command)
		return ""
	}
	ac.consumed[arg] = struct{}{}
	return ac.positionalArgs[arg]
}

func (ac *ArgConsumer) CheckAllConsumed() {
	if len(ac.consumed) < len(ac.positionalArgs) {
		extra := []string{}
		for i, arg := range ac.positionalArgs {
			if _, consumed := ac.consumed[i]; !consumed {
				extra = append(extra, arg)
			}
		}
		argumentInsert := "argument"
		if len(extra) > 1 {
			argumentInsert = "arguments"
		}
		ac.diagnose(fmt.Sprintf("Incorrect usage: invalid %s '%s'.", argumentInsert, strings.Join(extra, " ")), ac.command)
	}
}
