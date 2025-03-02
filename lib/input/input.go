package input

import (
	"bufio"
	"io"
	"os"
	"strings"

	"msh/lib/errco"
	"msh/lib/servctrl"
	"msh/lib/servstats"
)

// GetInput is used to read input from user.
// [goroutine]
func GetInput() {
	var line string
	var err error

	reader := bufio.NewReader(os.Stdin)

	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			// if stdin is unavailable (msh running as service)
			// exit from input goroutine to avoid an infinite loop
			if err == io.EOF {
				// in case input goroutine returns abnormally while msh is running in terminal,
				// the user must be notified with errco.LVL_B
				errco.LogMshErr(errco.NewErr(errco.ERROR_INPUT_UNAVAILABLE, errco.LVL_B, "GetInput", "stdin unavailable, exiting input goroutine"))
				return
			}
			errco.LogMshErr(errco.NewErr(errco.ERROR_INPUT_READ, errco.LVL_D, "GetInput", err.Error()))
			continue
		}

		// make sure that only 1 space separates words
		line = strings.ReplaceAll(line, "\n", "")
		line = strings.ReplaceAll(line, "\r", "")
		line = strings.ReplaceAll(line, "\t", " ")
		for strings.Contains(line, "  ") {
			line = strings.ReplaceAll(line, "  ", " ")
		}
		lineSplit := strings.Split(line, " ")

		errco.Logln(errco.LVL_D, "GetInput: user input: %s", lineSplit[:])

		switch lineSplit[0] {
		// target msh
		case "msh":
			// check that there is a command for the target
			if len(lineSplit) < 2 {
				errco.LogMshErr(errco.NewErr(errco.ERROR_COMMAND_INPUT, errco.LVL_A, "GetInput", "specify msh command (start - freeze - exit)"))
				continue
			}

			switch lineSplit[1] {
			case "start":
				errMsh := servctrl.StartMS()
				if errMsh != nil {
					errco.LogMshErr(errMsh.AddTrace("GetInput"))
				}
			case "freeze":
				// stop minecraft server with no player check
				errMsh := servctrl.StopMS(false)
				if errMsh != nil {
					errco.LogMshErr(errMsh.AddTrace("GetInput"))
				}
			case "exit":
				errMsh := servctrl.StopMS(false)
				if errMsh != nil {
					errco.LogMshErr(errMsh.AddTrace("GetInput"))
				}
				errco.Logln(errco.LVL_A, "exiting msh")
				os.Exit(0)
			default:
				errco.LogMshErr(errco.NewErr(errco.ERROR_COMMAND_UNKNOWN, errco.LVL_A, "GetInput", "unknown command (start - freeze - exit)"))
			}

		// taget minecraft server
		case "mine":
			// check that there is a command for the target
			if len(lineSplit) < 2 {
				errco.LogMshErr(errco.NewErr(errco.ERROR_COMMAND_INPUT, errco.LVL_A, "GetInput", "specify mine command"))
				continue
			}

			// check if server is online
			if servstats.Stats.Status != errco.SERVER_STATUS_ONLINE {
				errco.LogMshErr(errco.NewErr(errco.ERROR_SERVER_NOT_ONLINE, errco.LVL_A, "GetInput", "minecraft server is not online (try \"msh start\")"))
				continue
			}

			// pass the command to the minecraft server terminal
			_, errMsh := servctrl.Execute(strings.Join(lineSplit[1:], " "), "user input")
			if errMsh != nil {
				errco.LogMshErr(errMsh.AddTrace("GetInput"))
			}

		// wrong target
		default:
			errco.LogMshErr(errco.NewErr(errco.ERROR_COMMAND_INPUT, errco.LVL_A, "GetInput", "specify the target (msh - mine)"))
		}
	}
}
