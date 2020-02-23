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

This project designed to solve the problem of cleaning resources from the failed nodes that's blocks any further operation and recovery.

Fencing is neccesary if you want to have redundancy for your StatefulSet pods.

If any node falls, **kube-fencing** will guaranteed kill it via **fence-agent**, afterwards it will clear the node of all resources, that's make Kubernetes possible to schedule pods on the rest nodes.

Kube-fencing includes three containers:

### fencing-controller

The main controller which watches for the node states, and if one of them becomes to the `NotReady` due `NodeStatusUnknown` reason, runs fencing procedure.

### fencing-switcher

This is small container which can be deployed as daemonset, it will enable fencing during start, and disable fencing when node is gracefully shutdowns or reboots.

### fencing-agents

This container contains installed `fence-agents` package.

When fencing procedure is called **fecning-controller** creates Job which can use **fecning-agents** image to execute specific fencing agent.
If fencing was successful it will celanup (or delete) the node from the kubernetes.

The next fencing agents are included:

```
fence_ack_manual      fence_brocade         fence_dummy           fence_idrac           fence_ilo4_ssh        fence_ipmilan         fence_ovh             fence_rsb             fence_vmware          
fence_alom            fence_cisco_mds       fence_eaton_snmp      fence_ifmib           fence_ilo_moonshot    fence_ironic          fence_powerman        fence_sanbox2         fence_vmware_soap     
fence_amt             fence_cisco_ucs       fence_emerson         fence_ilo             fence_ilo_mp          fence_kdump           fence_pve             fence_sbd             fence_wti             
fence_apc             fence_compute         fence_eps             fence_ilo2            fence_ilo_ssh         fence_ldom            fence_raritan         fence_scsi            fence_xenapi          
fence_apc_snmp        fence_docker          fence_hds_cb          fence_ilo3            fence_imm             fence_lpar            fence_rcd_serial      fence_tripplite_snmp  fence_zvmip           
fence_azure_arm       fence_drac            fence_hpblade         fence_ilo3_ssh        fence_intelmodular    fence_mpath           fence_rhevm           fence_vbox            
fence_bladecenter     fence_drac5           fence_ibmblade        fence_ilo4            fence_ipdu            fence_netio           fence_rsa             fence_virsh           
```

## Quick Start

### Install kube-fencing
```bash
kubectl apply -f https://github.com/kvaps/kube-fencing/raw/master/deploy/kube-fencing.yaml
```

### Apply example PodTemplate

```bash
# Simple notify example (with after-hook)
kubectl apply -f https://github.com/kvaps/kube-fencing/raw/master/deploy/examples/after-hook.yaml

# HP iLO example
kubectl apply -f https://github.com/kvaps/kube-fencing/raw/master/deploy/examples/hp-ilo.yaml
```

### Prepare own fencing template

Prepare your own fencing PodTemplate using the examples above.

Fencing-controller will spawn this PodTemplate every time when node going to unknown state.  
It also appends `fencing/node` and `fencing/id` annotations to the pod, thus allows you to use this information in your fencing command.

The specified command must ends with `0` exit-code when fencing was successful and return `1` exit-code when failed.

You can create multiple PodTemplates for different nodes, but `fencing` will be used by default.

## Configuration parameters

All configuration is reduced to the specific annotations.

You can specify the needed annotations for specific node or commonly for PodTemplate, hovewer node annotations take precedence.

| Annotation | Description | Default  |
|:-|:-|:-|
| `fencing/enabled` | Fencing-switcher automatically sets this annotation to enable or disable fencing for the node. *(can be specified only for node, usually you don't need to configure it)*. | `false` |
| `fencing/id`      | Specify the device id which will be used to fence the node. | *same as node name* |
| `fencing/template`| Specify PodTemplate which be used to fence the node. | `fencing` |
| `fencing/mode`    | Specify cleanup mode for the node: <ul><li><code>none</code> - do nothing after succesful fencing.</li><li><code>flush</code> - remove all pods and volumeattachments from the node after succesful fencing.</li><li><code>delete</code> - remove the node after succesful fencing.</li></ul>  | `flush` |
| `fencing/after-hook` | Specific PodTemplate which will be spawned after successful fencing. | *unspecified* |
| `fencing/timeout` | Timeout in seconds to wait for the node recovery before starting fencing procedure. | `0` |
| `fencing/ttl`     | `TTLSecondsAfterFinished` parameter for the created Jobs. | `100` |
