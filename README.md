
# Machine Inventory as Kubernetes Objects (MIKO)

This project provides a set of kubernetes CRDs to keep a bare-metal
machine inventory in a kubernetes data plane.

Additionally controllers are provided that keep various indices
on those objects to lookup those objects for dedicated use cases
required for running a bare-metal datacenter.

Those indices can be consumed directly by embedding the
appropriate module and controller into an own controller manager
or by enabling an included index server that provides a simple
REST API for mapping an index request to an object name of
the approriate kubernetes object.

The project is based on the
[controller manager library](https://github.com/gardener/controller-manager-library)
used to offer dedicated embeddable controllers, index modules and
the index server.

Additionally an own controller manager executable is provided which
offers the controllers and servers to run an index servers.

## The CRDs

The project provides three CRDs for dedicated purposes:
- The Machine Inventory CRD ([`MachineInfo`](pkg/apis/machines/v1alpha1/machineinfo.go)) is used to store
  information of the configuration and features of a dedicated bare-metal machine
- The Base Board Management Controller CRD (BMC) ([`BaseBoardManagementController`](pkg/apis/machines/v1alpha1/bmcinfo.go)) 
  is used to store information for the Out-Of-Band area including the IPMI information.
- The Machine Type CRD ([`MachineType`](pkg/apis/machines/v1alpha1/machinetype.go))
  is used to store machine type information discoverable by MAC address prefixes
  assigned by a dedicated vendor for a dedicated type of machine. 

## The Components

The project provides several controller manager components that can be resused
to compose and own controller manager

### Controllers

The controllers provide an appropriate index in the shared environment of the
controller manager, that can be queried by any other controller.

- `pkg/controllers/machines`
  
  A controller providing a machine info index that can be used to identify a machine
  according to its MAC addresses or UUID.

- `pkg/controllers/bmc`

  A controller providing a BMC info index that can be used to identify the BMC
  information of a machine according to its MAC address in the OOB network or UUID.
  
- `pkg/controllers/types`

  A controller providing a machine type index that can be used to identify the
  type of a machine according to its MAC addresses.
  
### Modules

- `pkg/modules/serverindex`

  The serverindex module provides indices based on an index server.
  It offers the machine and BMC info cache. The indices are exported
  to the controller managers shared environment and can be accessed
  by any other controller (only this module OR the appropriate controllers
  should be used in a controller manager)

### Servers

- `pkg/servers/machineindexer`

  A simple http web server offering the embedded indices. Indices are 
  provided by dedicated go packages that can be embedded into a dedicated
  controller manager with anonymous imports. The following indices are provided:
  - Machine Info index (`pkg/servers/machineindexer/machineinfo`) (path `info`)
    based on query parameters `mac`and `uuid`.
  - BMC Info index (`pkg/servers/machineindexer/bmcinfo`) (path `bmc`)
   based on query parameters `mac`and `uuid`.
  - Machine Type index (`pkg/servers/machineindexer/machinetype`) (path `type`)
   based on query parameter `mac`.
