#!/bin/bash

virsh -c qemu+tcp://localhost/system pool-define-as default dir - - - - "/var/lib/libvirt/images"
virsh -c qemu+tcp://localhost/system pool-start default
virsh -c qemu+tcp://localhost/system net-start default
