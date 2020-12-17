package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dvob/vu/internal/virt"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

const envPrefix = "CIS_"

var (
	version        = "n/a"
	commit         = "n/a"
	virStoragePool string
	mgr            *virt.LibvirtManager
)

func newRootCmd() *cobra.Command {
	var (
		uri string
	)
	rootCmd := &cobra.Command{
		Use: "vu",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			// flags from environment
			cmd.Flags().VisitAll(func(f *flag.Flag) {
				varName := envPrefix + strings.ToUpper(f.Name)
				if val, ok := os.LookupEnv(varName); !f.Changed && ok {
					err := f.Value.Set(val)
					if err != nil {
						errExit(err)
					}
				}
			})

			// initialize libvirt manager
			mgr, err = virt.NewLibvirtManager(virStoragePool, uri)
			if err != nil {
				errExit(err)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			err := mgr.Close()
			if err != nil {
				errExit(err)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&uri, "uri", "unix:/var/run/libvirt/libvirt-sock", "Connection url for libvirtd. e.g. tcp:localhost:16509")
	rootCmd.AddCommand(
		newConfigCmd(),
		newImageCmd(),
		newCreateCmd(),
		newListCmd(),
		newStartCmd(),
		newShutdownCmd(),
		newRemoveCmd(),
		newVersionCmd(),
		newCompletionCmd(),
	)
	return rootCmd
}

func addPoolOption(fs *flag.FlagSet, pool *string) {
	fs.StringVar(pool, "pool", "default", "The storage pool to use")
}

func newCompletionCmd() *cobra.Command {
	var shell string
	cmd := &cobra.Command{
		Use:       "completion <shell>",
		ValidArgs: []string{"bash", "zsh"},
		Args:      cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shell = args[0]
			var err error
			switch shell {
			case "bash":
				err = newRootCmd().GenBashCompletion(os.Stdout)
			case "zsh":
				err = newRootCmd().GenZshCompletion(os.Stdout)
			default:
				errExit("unknown shell")
			}
			if err != nil {
				errExit(err)
			}
		},
	}
	return cmd
}

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("version", version)
			fmt.Println("commit", commit)
		},
	}
	return cmd
}

func main() {
	_ = newRootCmd().Execute()
}
