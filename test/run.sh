#!/bin/bash

docker run -it --rm -p 16509:16509 --privileged libvirtd libvirtd --listen
