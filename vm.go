package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/dvob/vu/internal/cloud-init"
	"github.com/dvob/vu/internal/virt"
	"github.com/spf13/cobra"
)

const configVolumePrefix = "vu_config_"

func newCreateCmd() *cobra.Command {
	var (
		name           string
		names          []string
		baseImage      string
		user           string
		sshAuthKeyFile string
		passwordHash   string
		vcpus          uint
		memory         ByteSize
		network        string
		cloudCfg       *cloudinit.Config
		configDir      []string
		networkParams  = &cloudinit.NetworkParameter{}
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

			sshAuthKey, err := ioutil.ReadFile(sshAuthKeyFile)
			if err != nil {
				errExit(err)
			}

			if len(configDir) > 0 {
				for _, dir := range configDir {
					fi, err := os.Stat(dir)
					if err != nil {
						errExit(err)
					}
					if !fi.IsDir() {
						errExit(dir, "is not a directory")
					}
				}
			}

			var failed = false
			var i int

			for i, name = range names {

				var iso io.Reader
				var configVolumeName = configVolumePrefix + name

				if len(configDir) > 0 {
					iso, err = cloudinit.CreateISOFromDir(configDir[i%len(configDir)])
				} else {
					cloudCfg = cloudinit.NewDefaultConfig(name, user, string(sshAuthKey), passwordHash, networkParams)
					iso, err = cloudCfg.CreateISO()
				}
				if err != nil {
					errExit("failed to create config iso:", err)
				}

				_, err := mgr.CreateISOVolume(configVolumeName, iso)
				if err != nil {
					errExit("failed to create config volume:", err)
				}

				vmCfg := &virt.VMConfig{
					Name:            name,
					BaseImageVolume: baseImage,
					ISOImageVolume:  configVolumeName,
					Memory:          uint(memory),
					VCPU:            vcpus,
					Network:         network,
					DiskSize:        uint64(diskSize),
				}

				err = mgr.Create(vmCfg)
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
	cmd.Flags().UintVar(&vcpus, "cpus", 1, "amount of cpus")
	cmd.Flags().StringVar(&network, "network", "default", "name of the network")
	cmd.Flags().StringSliceVar(&configDir, "dir", []string{}, "use cloud init config from directory. if you start multiple VMs simultaneously you can provide multiple directories")
	addPasswordHashOption(cmd.Flags(), &passwordHash)
	addSSHAuthKeyOption(cmd.Flags(), &sshAuthKeyFile)
	addSSHUserOption(cmd.Flags(), &user)
	addNetworkOptions(cmd.Flags(), networkParams)
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

				err = mgr.RemoveVolume(configVolumePrefix + name)
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
	var (
		all bool
		vms []string
		err error
	)
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list VMs",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			mgr.ListAllDetail()
			if all {
				vms, err = mgr.ListAll()
			} else {
				vms, err = mgr.List()
			}
			if err != nil {
				errExit(err)
			}
			for _, vm := range vms {
				fmt.Println(vm)
			}

		},
	}
	cmd.Flags().BoolVarP(&all, "all", "a", false, "show also domains not created by vu")
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
