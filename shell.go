package main

const bash_completion_func = `
__cis_resource_images()
{
    local cis_images
    if cis_images=$( cis image list 2>/dev/null); then
        COMPREPLY=( $( compgen -W "$cis_images" -- "$cur" ) )
    fi
}

__cis_resource_vms()
{
    local cis_vms
    if cis_vms=$( cis list 2>/dev/null); then
        COMPREPLY=( $( compgen -W "$cis_vms" -- "$cur" ) )
    fi
}

__cis_custom_func() {
    case ${last_command} in
        cis_image_remove)
            __cis_resource_images
            return
            ;;
        cis_start | cis_shutdown | cis_remove)
            __cis_resource_vms
            return
            ;;
        cis_create)
	    if [[ ${#nouns[@]} -gt 0 ]]; then
	        return
	    fi
	    __cis_resource_images
            return
            ;;
        *)
            ;;
    esac
}
`
