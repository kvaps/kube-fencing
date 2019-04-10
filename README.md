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

This project was designed to solve the problem of cleaning resources from the failed nodes that's blocks any further operation and recovery.

Fencing is neccesary operation if you want to have redundancy for your StatefulSets.

If any node falls, **kube-fencing** will kill it via **fence-agent**, afterwards it will clear the node of all resources, that's make Kubernetes possible to schedule pods on the rest nodes.

Kube-fencing includes three services:

### fencing-controller

The main service which watches for the node states, and if one of them becomes to the `NotReady` due `NodeStatusUnknown` reason, runs fencing procedure.

### fencing-switcher

This is small daemonset which enable fencing for the each node during start up, and disable it when node is gracefully shutdowns or reboots.

### fencing-agents

The container with installed `fence-agents` package.

When fencing procedure is called **fecning-controller** goes to the **fecning-agents** pod and executes custom script there.
If script was success it will celanup (or delete) the node in the kubernetes.

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

## Quick Start

### Create namespace and rbac roles
```
kubectl apply -f https://github.com/kvaps/kube-fencing/raw/master/examples/00-fencing-namespace.yaml
kubectl apply -f https://github.com/kvaps/kube-fencing/raw/master/examples/fencing-rbac.yaml
```

### Apply main services
```
kubectl apply -f https://github.com/kvaps/kube-fencing/raw/master/examples/fencing-controller.yaml
kubectl apply -f https://github.com/kvaps/kube-fencing/raw/master/examples/fencing-agents.yaml
kubectl apply -f https://github.com/kvaps/kube-fencing/raw/master/examples/fencing-scripts.yaml
kubectl apply -f https://github.com/kvaps/kube-fencing/raw/master/examples/fencing-switcher.yaml
```

### Prepare fencing-script

Then you should download [`fencing-scripts.yaml`](https://github.com/kvaps/kube-fencing/raw/master/examples/fencing-scripts.yaml) configmap and prepare your own script there.
This script must describe fencing procedure for your infrastructure.

Script takes first argument as **node name**, and should return **zero exit-code** if fencing was succesful and **non-zero exit-code** if fencing was failed.

Another words **fencing-controller** will call this script like:
```
kubectl exec -n fencing fencing-agents-577dff5bf8-fp5np /scripts/fence.sh <nodename>
```

In the example script you can see the next procedure:

* take first argument and save it into NAME variable
* write `<nodename>-ilo` into ILO vatiable
* run **fence_ilo** agent via `<nodename>-ilo`.

You can implement any logic there.

In case if you have many different devices in your ifrastructure you can run multiple **fencing-controllers** with different label-selectors for the nodes, or implement all logic in one script, it's your choose.

### Mark nodes as fencing enabled

```
kubectl label node <nodename> fencing=enabled
```

*Note: after this action **fencing-swithcer** will mark nodes as `fencing=enabled` and `fencing=disabled` automatically when node is boots up or shuting down gracefully*

## Configuration parameters

All configuration is reduced to the environment variables.

You can specify needed variables inside yaml file for each service.

### fencing-controller

* **FENCING_NODE_SELECTOR**

  Label for fencing enabled nodes, fencing-controller will watch for the nodes only with this label <br>
  *(example: `fencing=enabled`)*
  
* **FENCING_AGENT_SELECTOR**

  Agent pod selector, fencing-controller will run script inside this pod, it should be `Running` in the same namespace with fencing-controller *(example: `app=fencing-agents`)*
  
* **FENCING_SCRIPT**

  Script to call on agent pod *(example: `/scripts/fence.sh`)*
  
* **FLUSHING_MODE**

  * `delete` - to delete node from cluster.
  * `recreate` - to flush all resources from node, and recreate node.
  * `none` - do nothing, just call the script.
  * `info` - do nothing and disable fencing call (for debugging)

* **DEBUG**

  If set, debug output will be enabled *(example: `1`)*
  

### fencing-switcher

* **FENCING_LABEL**

  Label name to switch on the node, this label will switch between `enabled`/`disabled` parameters *(example: `fencing`)*
  
* **NODE_NAME**

  Should always be equal `spec.nodeName` for the node where it is running.

* **DEBUG**

  If set, debug output will be enabled *(example: `1`)*
