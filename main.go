package main

import (
	"fmt"
	"os"
	"strings"

	vu "github.com/dvob/vu/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	envVarPrefix = "VU_"
)

var (
	version = "n/a"
	commit  = "n/a"
)

func main() {
	err := newRootCmd().Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var (
		mgr = &vu.Manager{}
	)
	opts := vu.NewLibvirtDefaultOptions()
	cmd := &cobra.Command{
		Use:              "vu",
		Short:            "vu spins up virtual machines using cloud-init images",
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				optName := strings.ToUpper(f.Name)
				optName = strings.ReplaceAll(optName, "-", "_")
				varName := envVarPrefix + optName
				if val, ok := os.LookupEnv(varName); ok {
					innerErr := f.Value.Set(val)
					if innerErr != nil {
						err = fmt.Errorf("invalid environment variable %s: %w", varName, innerErr)
					}
				}
			})
			if err != nil {
				return err
			}

			m, err := vu.NewLibvirtManager(opts)
			if err != nil {
				return err
			}
			*mgr = *m

			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			return err
		},
	}
	cmd.AddCommand(
		newImageCmd(mgr),
		newCreateCmd(mgr),
		newStartCmd(mgr),
		newShutdownCmd(mgr),
		newRemoveCmd(mgr),
		newListCmd(mgr),
		newConfigCmd(),
		newCompletionCmd(),
		newVersionCmd(),
	)
	return cmd
}

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(version, commit)
			return nil
		},
	}
	return cmd
}

func newCompletionCmd() *cobra.Command {
	var shell string
	cmd := &cobra.Command{
		Use:       "completion <shell>",
		ValidArgs: []string{"bash", "zsh", "fish", "ps"},
		Args:      cobra.ExactArgs(1),
		Hidden:    true,
		RunE: func(cmd *cobra.Command, args []string) error {
			shell = args[0]
			var err error
			switch shell {
			case "bash":
				err = newRootCmd().GenBashCompletion(os.Stdout)
			case "zsh":
				err = newRootCmd().GenZshCompletion(os.Stdout)
			case "fish":
				err = newRootCmd().GenFishCompletion(os.Stdout, true)
			case "ps":
				err = newRootCmd().GenPowerShellCompletion(os.Stdout)
			default:
				err = fmt.Errorf("unknown shell: %s", shell)
			}
			return err
		},
	}
	return cmd
}
