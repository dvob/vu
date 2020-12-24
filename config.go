package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/dvob/vu/internal/cloudinit"
	"github.com/spf13/cobra"
)

type cloudInitOptions struct {
	config *cloudinit.Config

	name          string
	user          string
	sshPubKey     string
	sshPubKeyFile string
	passwordHash  string

	ip      string
	gateway string
	dns     []string

	profiles []string
	dirs     []string
}

func getAbsProfileDirs(profiles []string) ([]string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve profile dirs: %w", err)
	}

	profileBaseDir := filepath.Join(userHome, ".config", "vu", "profiles")

	dirs := []string{}
	for _, profile := range profiles {
		dirs = append(dirs, filepath.Join(profileBaseDir, profile))
	}
	return dirs, nil
}

func (o *cloudInitOptions) complete() error {
	if o.user == "" {
		localUser, err := user.Current()
		if err != nil {
			return err
		}
		o.user = localUser.Username
	}
	if o.sshPubKey == "" {
		if o.sshPubKeyFile == "" {
			userHome, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			o.sshPubKeyFile = filepath.Join(userHome, ".ssh", "id_rsa.pub")
		}

		pubKey, err := ioutil.ReadFile(o.sshPubKeyFile)
		if err != nil {
			return err
		}
		o.sshPubKey = string(pubKey)
	}

	profileDirs, err := getAbsProfileDirs(o.profiles)
	if err != nil {
		return err
	}

	o.config, err = cloudinit.ConfigFromDir(profileDirs...)
	if err != nil {
		return err
	}

	dirConfig, err := cloudinit.ConfigFromDir(o.dirs...)
	if err != nil {
		return err
	}
	err = o.config.Merge(dirConfig)
	if err != nil {
		return err
	}

	if o.config.UserData == nil {
		o.config.UserData = &cloudinit.UserData{}
	}
	if o.config.MetaData == nil {
		o.config.MetaData = &cloudinit.MetaData{}
	}
	c := cloudinit.NewDefaultConfig(o.name, o.user, o.sshPubKey)
	return o.config.Merge(c)
}

func (o *cloudInitOptions) bindFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.user, "user", "", "user name to create in the during startup")
	cmd.Flags().StringVar(&o.sshPubKey, "ssh-pub-key", "", "ssh public key to use")
	cmd.Flags().StringVar(&o.passwordHash, "password-hash", "", "Password hash to login without SSH over console. The hash can be generated with openssl passwd.")
	cmd.Flags().StringVar(&o.ip, "ip", "", "configure static IPv4 address insted of DHCP. address has to be specified in CIDR notation.")
	cmd.Flags().StringVar(&o.gateway, "gateway", "", "the default IPv4 gateway. if no gateway is configured the lowest IP")
	cmd.Flags().StringSliceVar(&o.dns, "nameserver", []string{}, "configure the dns server address. if no address is configured the default gateway is used.")
	cmd.Flags().StringSliceVar(&o.profiles, "profile", []string{}, "base profile which want to use for our configuration")
	cmd.Flags().StringSliceVar(&o.dirs, "dir", []string{}, "use configuration from directory as base configuration")
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
	o := &cloudInitOptions{}
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "shows cloud init configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.name = args[0]
			err := o.complete()
			if err != nil {
				return err
			}

			output, err := o.config.String()
			if err != nil {
				return err
			}
			fmt.Println(output)
			return nil
		},
	}
	o.bindFlags(cmd)
	return cmd
}

func newConfigWriteCmd() *cobra.Command {
	var (
		o      = &cloudInitOptions{}
		target string
		iso    bool
	)
	cmd := &cobra.Command{
		Use:   "write NAME TARGET",
		Short: "writes cloud init config to directory or iso file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.name = args[0]
			target = args[1]

			err := o.complete()
			if err != nil {
				return err
			}

			if iso {
				isoData, err := o.config.ISO()
				if err != nil {
					return err
				}
				return ioutil.WriteFile(target, isoData, 0640)
			}
			return o.config.ToDir(target)
		},
	}
	cmd.Flags().BoolVar(&iso, "iso", false, "write cloud init config to iso file")
	o.bindFlags(cmd)
	return cmd
}
