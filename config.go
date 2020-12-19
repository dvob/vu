package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/dvob/vu/internal/cloud-init"
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

func addNetworkOptions(fs *flag.FlagSet, np *cloudinit.NetworkParameter) {
	fs.StringVar(&np.Address, "ip", "", "configure static IPv4 address insted of DHCP. address has to be specified in CIDR notation.")
	fs.StringVar(&np.Gateway, "gateway", "", "the default IPv4 gateway. if no gateway is configured the lowest IP")
	fs.StringSliceVar(&np.Nameserver, "nameserver", []string{}, "configure the dns server address. if no address is configured the default gateway is used.")
}

func addPasswordHashOption(fs *flag.FlagSet, dest *string) {
	fs.StringVar(dest, "password-hash", "", "Password hash to login without SSH over console. The hash can be generated with openssl passwd.")
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
		passwordHash   string
		networkParams  = &cloudinit.NetworkParameter{}
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
			cfg := cloudinit.NewDefaultConfig(name, user, string(sshAuthKey), passwordHash, networkParams)
			if err != nil {
				errExit(err)
			}

			cfgOut, err := cfg.String()
			if err != nil {
				errExit(err)
			}
			fmt.Println(cfgOut)
		},
	}
	addSSHUserOption(cmd.Flags(), &user)
	addSSHAuthKeyOption(cmd.Flags(), &sshAuthKeyFile)
	addPasswordHashOption(cmd.Flags(), &passwordHash)
	addNetworkOptions(cmd.Flags(), networkParams)
	return cmd
}

func newConfigWriteCmd() *cobra.Command {
	var (
		name           string
		user           string
		sshAuthKeyFile string
		passwordHash   string
		networkParams  = &cloudinit.NetworkParameter{}
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
			cfg := cloudinit.NewDefaultConfig(name, user, string(sshAuthKey), passwordHash, networkParams)
			err = cfg.WriteToDir(dest)
			if err != nil {
				errExit(err)
			}
		},
	}
	addSSHUserOption(cmd.Flags(), &user)
	addSSHAuthKeyOption(cmd.Flags(), &sshAuthKeyFile)
	addPasswordHashOption(cmd.Flags(), &passwordHash)
	addNetworkOptions(cmd.Flags(), networkParams)
	return cmd
}
