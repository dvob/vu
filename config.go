package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/dsbrng25b/cis/internal/cloud-init"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func addSSHUserOption(fs *flag.FlagSet, dest *string) {
	var userName string
	localUser, err := user.Current()
	if err != nil {
		userName = "root"
	} else {
		userName = localUser.Username
	}

	fs.StringVar(dest, "ssh-user", userName, "user name to create in the during startup")
}

func addSSHAuthKeyOption(fs *flag.FlagSet, dest *string) {
	userHome, _ := os.UserHomeDir()
	path := filepath.Join(userHome, ".ssh", "id_rsa.pub")
	fs.StringVar(dest, "ssh-auth-key", path, "ssh public key to use")
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "helps to debug cloud init configuration",
	}
	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigWriteCmd(),
	)
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	var (
		name           string
		user           string
		sshAuthKeyFile string
	)
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "shows cloud init configuration",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]

			sshAuthKey, err := ioutil.ReadFile(sshAuthKeyFile)
			if err != nil {
				errExit(err)
			}
			cfg := cloudinit.NewDefaultConfig(name, user, string(sshAuthKey))
			cfgOut, err := cfg.String()
			if err != nil {
				errExit(err)
			}
			fmt.Println(cfgOut)
		},
	}
	addSSHUserOption(cmd.Flags(), &user)
	addSSHAuthKeyOption(cmd.Flags(), &sshAuthKeyFile)
	return cmd
}

func newConfigWriteCmd() *cobra.Command {
	var (
		name           string
		user           string
		sshAuthKeyFile string
		dest           string
	)
	cmd := &cobra.Command{
		Use:   "write <name> <dest-dir>",
		Short: "writes cloud init config to directory",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]
			dest = args[1]

			sshAuthKey, err := ioutil.ReadFile(sshAuthKeyFile)
			if err != nil {
				errExit(err)
			}
			cfg := cloudinit.NewDefaultConfig(name, user, string(sshAuthKey))
			err = cfg.WriteToDir(dest)
			if err != nil {
				errExit(err)
			}
		},
	}
	addSSHUserOption(cmd.Flags(), &user)
	addSSHAuthKeyOption(cmd.Flags(), &sshAuthKeyFile)
	return cmd
}
