package command

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/cli"
)

// ServeCommand is a Command implementation that applies a Terraform
// configuration and actually builds or changes infrastructure.
type ServeCommand struct {
	Meta
}

func (c *ServeCommand) commands(w io.Writer) map[string]cli.CommandFactory {
	meta := c.Meta
	meta.oldUi = nil
	meta.Ui = &cli.ConcurrentUi{
		Ui: &ColorizeUi{
			Colorize:   meta.Colorize(),
			ErrorColor: "[red]",
			WarnColor:  "[yellow]",
			Ui: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      w,
				ErrorWriter: w,
			},
		},
	}

	cmds := map[string]cli.CommandFactory{
		"apply": func() (cli.Command, error) {
			return &ApplyCommand{
				Meta: meta,
			}, nil
		},

		"console": func() (cli.Command, error) {
			return &ConsoleCommand{
				Meta: meta,
			}, nil
		},

		"destroy": func() (cli.Command, error) {
			return &ApplyCommand{
				Meta:    meta,
				Destroy: true,
			}, nil
		},

		"env": func() (cli.Command, error) {
			return &WorkspaceCommand{
				Meta:       meta,
				LegacyName: true,
			}, nil
		},

		"env list": func() (cli.Command, error) {
			return &WorkspaceListCommand{
				Meta:       meta,
				LegacyName: true,
			}, nil
		},

		"env select": func() (cli.Command, error) {
			return &WorkspaceSelectCommand{
				Meta:       meta,
				LegacyName: true,
			}, nil
		},

		"env new": func() (cli.Command, error) {
			return &WorkspaceNewCommand{
				Meta:       meta,
				LegacyName: true,
			}, nil
		},

		"env delete": func() (cli.Command, error) {
			return &WorkspaceDeleteCommand{
				Meta:       meta,
				LegacyName: true,
			}, nil
		},

		"fmt": func() (cli.Command, error) {
			return &FmtCommand{
				Meta: meta,
			}, nil
		},

		"get": func() (cli.Command, error) {
			return &GetCommand{
				Meta: meta,
			}, nil
		},

		"graph": func() (cli.Command, error) {
			return &GraphCommand{
				Meta: meta,
			}, nil
		},

		"import": func() (cli.Command, error) {
			return &ImportCommand{
				Meta: meta,
			}, nil
		},

		"init": func() (cli.Command, error) {
			return &InitCommand{
				Meta: meta,
			}, nil
		},

		"internal-plugin": func() (cli.Command, error) {
			return &InternalPluginCommand{
				Meta: meta,
			}, nil
		},

		// "terraform login" is disabled until Terraform Cloud is ready to
		// support it.
		/*
			"login": func() (cli.Command, error) {
				return &LoginCommand{
					Meta: meta,
				}, nil
			},
		*/

		"output": func() (cli.Command, error) {
			return &OutputCommand{
				Meta: meta,
			}, nil
		},

		"plan": func() (cli.Command, error) {
			return &PlanCommand{
				Meta: meta,
			}, nil
		},

		"providers": func() (cli.Command, error) {
			return &ProvidersCommand{
				Meta: meta,
			}, nil
		},

		"providers schema": func() (cli.Command, error) {
			return &ProvidersSchemaCommand{
				Meta: meta,
			}, nil
		},

		"push": func() (cli.Command, error) {
			return &PushCommand{
				Meta: meta,
			}, nil
		},

		"refresh": func() (cli.Command, error) {
			return &RefreshCommand{
				Meta: meta,
			}, nil
		},

		"show": func() (cli.Command, error) {
			return &ShowCommand{
				Meta: meta,
			}, nil
		},

		"taint": func() (cli.Command, error) {
			return &TaintCommand{
				Meta: meta,
			}, nil
		},

		"validate": func() (cli.Command, error) {
			return &ValidateCommand{
				Meta: meta,
			}, nil
		},

		"untaint": func() (cli.Command, error) {
			return &UntaintCommand{
				Meta: meta,
			}, nil
		},

		"workspace": func() (cli.Command, error) {
			return &WorkspaceCommand{
				Meta: meta,
			}, nil
		},

		"workspace list": func() (cli.Command, error) {
			return &WorkspaceListCommand{
				Meta: meta,
			}, nil
		},

		"workspace select": func() (cli.Command, error) {
			return &WorkspaceSelectCommand{
				Meta: meta,
			}, nil
		},

		"workspace show": func() (cli.Command, error) {
			return &WorkspaceShowCommand{
				Meta: meta,
			}, nil
		},

		"workspace new": func() (cli.Command, error) {
			return &WorkspaceNewCommand{
				Meta: meta,
			}, nil
		},

		"workspace delete": func() (cli.Command, error) {
			return &WorkspaceDeleteCommand{
				Meta: meta,
			}, nil
		},

		//-----------------------------------------------------------
		// Plumbing
		//-----------------------------------------------------------

		"0.12upgrade": func() (cli.Command, error) {
			return &ZeroTwelveUpgradeCommand{
				Meta: meta,
			}, nil
		},

		"debug": func() (cli.Command, error) {
			return &DebugCommand{
				Meta: meta,
			}, nil
		},

		"debug json2dot": func() (cli.Command, error) {
			return &DebugJSON2DotCommand{
				Meta: meta,
			}, nil
		},

		"force-unlock": func() (cli.Command, error) {
			return &UnlockCommand{
				Meta: meta,
			}, nil
		},

		"state": func() (cli.Command, error) {
			return &StateCommand{}, nil
		},

		"state list": func() (cli.Command, error) {
			return &StateListCommand{
				Meta: meta,
			}, nil
		},

		"state rm": func() (cli.Command, error) {
			return &StateRmCommand{
				StateMeta: StateMeta{
					Meta: meta,
				},
			}, nil
		},

		"state mv": func() (cli.Command, error) {
			return &StateMvCommand{
				StateMeta: StateMeta{
					Meta: meta,
				},
			}, nil
		},

		"state pull": func() (cli.Command, error) {
			return &StatePullCommand{
				Meta: meta,
			}, nil
		},

		"state push": func() (cli.Command, error) {
			return &StatePushCommand{
				Meta: meta,
			}, nil
		},

		"state show": func() (cli.Command, error) {
			return &StateShowCommand{
				Meta: meta,
			}, nil
		},
	}
	return cmds
}

func (c *ServeCommand) Run(args []string) int {
	var port int
	args, err := c.Meta.process(args, true)
	if err != nil {
		return 1
	}

	cmdName := "apply"

	cmdFlags := c.Meta.extendedFlagSet(cmdName)
	cmdFlags.IntVar(&port, "port", 8080, "listen")
	cmdFlags.Usage = func() { c.Ui.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: &handler{
			ServeCommand: c,
		},
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf(err.Error())
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		// handle err
		log.Printf(err.Error())
	}

	return 0
}

type handler struct {
	*ServeCommand
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmd_args, _ := r.URL.Query()["args"]
	log.Printf("[INFO] CLI command args: %#v", cmd_args)

	cliRunner := &cli.CLI{
		Args:       cmd_args,
		Commands:   h.ServeCommand.commands(w),
		HelpWriter: w,
	}

	exitCode, err := cliRunner.Run()
	if err != nil {
		log.Printf("[ERROR] Error executing CLI: %s, ExitCode: %d",
			err.Error(),
			exitCode)
	}
}

func (c *ServeCommand) Help() string {
	return c.help()
}

func (c *ServeCommand) help() string {
	helpText := `
Usage: terraform serve [options]

Options:

  -port=8080             Port to listen, Defaults to be 8080
`
	return strings.TrimSpace(helpText)
}

func (c *ServeCommand) Synopsis() string {
	return "Serve terraform on http"
}
