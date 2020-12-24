package main

import (
	"fmt"

	vu "github.com/dvob/vu/internal"
	"github.com/dvob/vu/internal/cloudinit"
	"github.com/dvob/vu/internal/vm"
	"github.com/spf13/cobra"
)

type vmOptions struct {
	vm vm.Config
	ci cloudInitOptions
}

func (o *vmOptions) complete() error {
	err := o.ci.complete()
	if err != nil {
		return err
	}

	return nil
}

func (o *vmOptions) bindFlags(cmd *cobra.Command) {
	o.ci.bindFlags(cmd)

	cmd.Flags().Var(NewByteSize(&o.vm.Memory), "memory", "amount of memory")
	cmd.Flags().Var(NewByteSize(&o.vm.DiskSize), "disk-size", "size of the cloned image")

	cmd.Flags().UintVar(&o.vm.CPUCount, "cpu", 1, "number of vCPUs")
	cmd.Flags().StringVar(&o.vm.Network, "network", "default", "name of the network to connect to")
}

func newCreateCmd(mgr *vu.Manager) *cobra.Command {
	options := &vmOptions{
		vm: vm.Config{
			// 1Gib
			Memory: 1_073_741_824,
		},
	}
	cmd := &cobra.Command{
		Use:   "create BASE_IMAGE NAME...",
		Short: "create new VMs from a base image",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := options.complete()
			if err != nil {
				return err
			}
			baseImage := args[0]
			names := args[1:]

			for _, name := range names {
				nameConfig := cloudinit.NewDefaultConfig(name, options.ci.user, options.ci.sshPubKey)

				// TODO: copy?
				err := options.ci.config.Merge(nameConfig)
				if err != nil {
					return err
				}

				err = mgr.Create(name, baseImage, &options.vm, options.ci.config)
				if err != nil {
					return err
				}
			}
			return nil
		},
		ValidArgsFunction: completeBaseImageFunc(mgr),
	}
	options.bindFlags(cmd)
	return cmd
}

func newRemoveCmd(mgr *vu.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove NAME...",
		Short:   "remove VMs",
		Aliases: []string{"rm"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			names := args

			for _, name := range names {
				err := mgr.Remove(name)
				if err != nil {
					return err
				}
			}

			return nil
		},
		ValidArgsFunction: completeVMFunc(mgr),
	}
	return cmd
}

func newStartCmd(mgr *vu.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start NAME...",
		Short: "starts VMs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			names := args
			for _, name := range names {
				err := mgr.VM.Start(name)
				if err != nil {
					return err
				}
			}
			return nil
		},
		ValidArgsFunction: completeVMFunc(mgr),
	}
	return cmd
}

func newListCmd(mgr *vu.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list VMs",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			vms, err := mgr.VM.List()
			if err != nil {
				return err
			}
			for _, vm := range vms {
				fmt.Println(vm.Name)
			}
			return nil
		},
	}
	return cmd
}

func newShutdownCmd(mgr *vu.Manager) *cobra.Command {
	var (
		force bool
	)
	cmd := &cobra.Command{
		Use:   "shutdown NAME...",
		Short: "shutdown VMs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			names := args
			for _, name := range names {
				err := mgr.VM.Shutdown(name, force)
				if err != nil {
					return err
				}
			}
			return nil
		},
		ValidArgsFunction: completeVMFunc(mgr),
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force shutdown")
	return cmd
}

func completeVMFunc(mgr *vu.Manager) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		vms, err := mgr.VM.List()
		if err != nil {
			fmt.Println(err)
		}
		vmNames := []string{}
		for _, vm := range vms {
			vmNames = append(vmNames, vm.Name)
		}
		return vmNames, cobra.ShellCompDirectiveNoFileComp
	}
}
