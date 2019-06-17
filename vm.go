package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dsbrng25b/cis/internal/cloud-init"
	"github.com/dsbrng25b/cis/internal/virt"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var (
		name           string
		names          []string
		baseImage      string
		user           string
		sshAuthKeyFile string
		passwordHash   string
		vcpus          int
		memory         ByteSize
		network        string
		vmCfg          *virt.VMConfig
		cloudCfg       *cloudinit.Config
		diskSize       ByteSize
	)

	// set default
	_ = memory.Set("1024m")
	cmd := &cobra.Command{
		Use:   "create <base_image> <name>...",
		Short: "create new VMs from base image",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			baseImage = args[0]
			names = args[1:]

			vmCfg = &virt.VMConfig{
				BaseImageVolume: baseImage,
				Memory:          uint(memory),
				VCPU:            vcpus,
				Network:         network,
				DiskSize:        uint64(diskSize),
			}

			sshAuthKey, err := ioutil.ReadFile(sshAuthKeyFile)
			if err != nil {
				errExit(err)
			}

			failed := false
			for _, name = range names {
				cloudCfg = cloudinit.NewDefaultConfig(name, user, string(sshAuthKey), passwordHash)
				err = mgr.Create(name, vmCfg, cloudCfg)
				if err != nil {
					failed = true
					errPrint(name+":", err)
				}
			}
			if failed {
				os.Exit(1)
			}

		},
	}
	cmd.Flags().Var(&memory, "memory", "amount of memory")
	cmd.Flags().Var(&diskSize, "disk-size", "size of the cloned image")
	cmd.Flags().IntVar(&vcpus, "cpus", 1, "amount of cpus")
	cmd.Flags().StringVar(&network, "network", "default", "name of the network")
	addPasswordHashOption(cmd.Flags(), &passwordHash)
	addSSHAuthKeyOption(cmd.Flags(), &sshAuthKeyFile)
	addSSHUserOption(cmd.Flags(), &user)
	return cmd
}

func newRemoveCmd() *cobra.Command {
	var (
		name  string
		names []string
	)
	cmd := &cobra.Command{
		Use:     "remove <name>...",
		Short:   "remove VMs",
		Aliases: []string{"rm"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			names = args

			failed := false
			for _, name = range names {
				err := mgr.Remove(name)
				if err != nil {
					failed = true
					errPrint(err)
				}
			}
			if failed {
				os.Exit(1)
			}
		},
	}
	return cmd
}

func newStartCmd() *cobra.Command {
	var (
		name  string
		names []string
	)
	cmd := &cobra.Command{
		Use:   "start <name>...",
		Short: "starts VMs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			names = args

			failed := false
			for _, name = range names {
				err := mgr.Start(name)
				if err != nil {
					failed = true
					errPrint(err)
				}
			}
			if failed {
				os.Exit(1)
			}
		},
	}
	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list VMs",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			vms, err := mgr.List()
			if err != nil {
				errExit(err)
			}
			for _, vm := range vms {
				fmt.Println(vm)
			}

		},
	}
	return cmd
}

func newShutdownCmd() *cobra.Command {
	var (
		name  string
		names []string
		force bool
	)
	cmd := &cobra.Command{
		Use:   "shutdown <name>...",
		Short: "shutdown VMs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			names = args

			failed := false
			for _, name = range names {
				err := mgr.Shutdown(name, force)
				if err != nil {
					failed = true
					errPrint(err)
				}
			}
			if failed {
				os.Exit(1)
			}
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force shutdown")
	return cmd
}
