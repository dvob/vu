package main

import (
	"fmt"
	"github.com/dsbrng25b/cis/internal/virt"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"os"
)

var version = "n/a"
var gitCommit = "n/a"
var buildTime = "n/a"

var virStoragePool string
var virConnectUrl string
var mgr *virt.LibvirtManager

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "cis",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			mgr, err = virt.NewLibvirtManager(virStoragePool, virConnectUrl)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(
		newConfigCmd(),
		newImageCmd(),
		newCreateCmd(),
		newStartCmd(),
		newStopCmd(),
		newRemoveCmd(),
		newVersionCmd(),
	)
	return rootCmd
}

func addPoolOption(fs *flag.FlagSet, pool *string) {
	fs.StringVar(pool, "pool", "default", "The storage pool to use")
}

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version and build information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("version:", version)
			fmt.Println("commit:", gitCommit)
			fmt.Println("build time:", buildTime)
		},
	}
	return cmd
}

func main() {
	newRootCmd().Execute()
}
