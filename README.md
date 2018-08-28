# kube-fencing

Fencing implementation for Kubernetes


| Image                          | Build Status                       |
|--------------------------------|------------------------------------|
| **[kube-fencing-controller]**  | ![kube-fencing-controller-status]  |
| **[kube-fencing-switcher]**    | ![kube-fencing-switcher-status]    |
| **[kube-fencing-agents]**      | ![kube-fencing-agents-status]      |


[kube-fencing-controller]: https://hub.docker.com/r/kvaps/kube-fencing-controller/
[kube-fencing-switcher]: https://hub.docker.com/r/kvaps/kube-fencing-switcher/
[kube-fencing-agents]: https://hub.docker.com/r/kvaps/kube-fencing-agents/
[kube-fencing-controller-status]: https://img.shields.io/docker/build/kvaps/kube-fencing-controller.svg
[kube-fencing-switcher-status]:  https://img.shields.io/docker/build/kvaps/kube-fencing-switcher.svg
[kube-fencing-agents-status]:  https://img.shields.io/docker/build/kvaps/kube-fencing-agents.svg


## Overview

This project was designed to solve the problem of cleaning resources from failed nodes that's blocks further operation and recovery.

Fencing is neccesary operation if you want to have redundancy for your StatefulSets.

If any node falls, kube-fencing will kill it via fence-agent, afterwards it will clear the node of all resources, that's make it possible to run them on the rest nodes.

Kube-fencing includes three services:

### fencing-controller

The main service which watches for the node states, and if one of them becomes to the `NotReady` due `NodeStatusUnknown` reason, runs fencing procedure.

### fencing-switcher

This is small daemonset which enable fencing for the node during start up, and disables it when it gracefully shutdown or reboots.

### fencing-agents

The container with installed `fence-agents` package.

When fencing procedure is needed **fecning-controller** goes to the **fecning-agents** pod and executes custom script there.
If script was success and finished with 0 exit code it will celanup (or delete) the node in kubernetes.

The next fencing agetns are included:

```
fence_ack_manual      fence_brocade         fence_dummy           fence_idrac           fence_ilo4_ssh        fence_ipmilan         fence_ovh             fence_rsb             fence_vmware          
fence_alom            fence_cisco_mds       fence_eaton_snmp      fence_ifmib           fence_ilo_moonshot    fence_ironic          fence_powerman        fence_sanbox2         fence_vmware_soap     
fence_amt             fence_cisco_ucs       fence_emerson         fence_ilo             fence_ilo_mp          fence_kdump           fence_pve             fence_sbd             fence_wti             
fence_apc             fence_compute         fence_eps             fence_ilo2            fence_ilo_ssh         fence_ldom            fence_raritan         fence_scsi            fence_xenapi          
fence_apc_snmp        fence_docker          fence_hds_cb          fence_ilo3            fence_imm             fence_lpar            fence_rcd_serial      fence_tripplite_snmp  fence_zvmip           
fence_azure_arm       fence_drac            fence_hpblade         fence_ilo3_ssh        fence_intelmodular    fence_mpath           fence_rhevm           fence_vbox            
fence_bladecenter     fence_drac5           fence_ibmblade        fence_ilo4            fence_ipdu            fence_netio           fence_rsa             fence_virsh           
```
