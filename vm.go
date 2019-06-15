package main

import (
	"fmt"
	"github.com/dsbrng25b/cis/internal/cloud-init"
	"github.com/dsbrng25b/cis/internal/virt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

func newCreateCmd() *cobra.Command {
	var (
		name           string
		baseImage      string
		user           string
		sshAuthKeyFile string
		vcpus          int
		memory         = ByteSize(1000 * 1000 * 1024)
		network        string
		vmCfg          *virt.VMConfig
		cloudCfg       *cloudinit.Config
	)
	cmd := &cobra.Command{
		Use:   "create <name> <base_image>",
		Short: "create a new VM from base image",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]
			baseImage = args[1]

			vmCfg = virt.NewDefaultVMConfig(name, baseImage)
			vmCfg.Memory = uint(memory)
			vmCfg.VCPU = vcpus
			vmCfg.Network = network

			sshAuthKey, err := ioutil.ReadFile(sshAuthKeyFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			cloudCfg = cloudinit.NewDefaultConfig(name, user, string(sshAuthKey))

			err = mgr.Create(name, vmCfg, cloudCfg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		},
	}
	cmd.Flags().Var(&memory, "memory", "amount of memory")
	cmd.Flags().IntVar(&vcpus, "cpus", 1, "amount of cpus")
	cmd.Flags().StringVar(&network, "network", "default", "name of the network")
	addSSHAuthKeyOption(cmd.Flags(), &sshAuthKeyFile)
	addSSHUserOption(cmd.Flags(), &user)
	return cmd
}

func newRemoveCmd() *cobra.Command {
	var (
		name string
	)
	cmd := &cobra.Command{
		Use:     "remove <name>",
		Short:   "removes VM",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]

			err := mgr.Remove(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		},
	}
	return cmd
}

func newStartCmd() *cobra.Command {
	var (
		name string
	)
	cmd := &cobra.Command{
		Use:   "start <name>",
		Short: "starts VM",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]

			err := mgr.Start(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		},
	}
	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <name>",
		Short:   "list VMs",
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			vms, err := mgr.List()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, vm := range vms {
				fmt.Println(vm)
			}

		},
	}
	return cmd
}

func newStopCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "stop <name>",
		Short: "stop VM",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]

			err := mgr.Stop(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	return cmd
}
