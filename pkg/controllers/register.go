/*
 * Copyright (c) 2020 by The metal-stack Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package controllers

import (
	"sync"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"

	"github.com/onmetal/k8s-machines/pkg/machines"
)

////////////////////////////////////////////////////////////////////////////////

var key = ctxutil.SimpleKey("machineindex")

func GetOrCreateMachineIndex(env extension.Environment, indexcreator func() machines.MachineIndex) machines.MachineIndex {
	return env.ControllerManager().GetOrCreateSharedValue(key, func() interface{} {
		return indexcreator()
	}).(machines.MachineIndex)
}

func GetMachineIndex(env extension.Environment) machines.MachineIndex {
	i := env.ControllerManager().GetSharedValue(key)
	if i != nil {
		return i.(machines.MachineIndex)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

var bmckey = ctxutil.SimpleKey("bmcindex")

func GetOrCreateBMCIndex(env extension.Environment, indexcreator func() machines.BMCIndex) machines.BMCIndex {
	return env.ControllerManager().GetOrCreateSharedValue(bmckey, func() interface{} {
		return indexcreator()
	}).(machines.BMCIndex)
}

func GetBMCIndex(env extension.Environment) machines.BMCIndex {
	i := env.ControllerManager().GetSharedValue(bmckey)
	if i != nil {
		return i.(machines.BMCIndex)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type Client interface{}

type MachineClient interface {
	PropagateMachineIndex(machines.MachineIndex)
}

type BMCClient interface {
	PropagateBMCIndex(machines.BMCIndex)
}

type registry struct {
	lock           sync.Mutex
	clients        []Client
	machineindices []machines.MachineIndex
	bmcindices     []machines.BMCIndex
}

var defaultRegistry = &registry{}

func RegisterClient(c Client) {
	defaultRegistry.RegisterClient(c)
}

func (this *registry) RegisterClient(c Client) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.clients = append(this.clients, c)
	if p, ok := c.(MachineClient); ok {
		for _, e := range this.machineindices {
			p.PropagateMachineIndex(e)
		}
	}
	if p, ok := c.(BMCClient); ok {
		for _, e := range this.bmcindices {
			p.PropagateBMCIndex(e)
		}
	}
}

func PropagateMachineIndex(index machines.MachineIndex) {
	defaultRegistry.PropagateMachineIndex(index)
}

func PropagateBMCIndex(index machines.BMCIndex) {
	defaultRegistry.PropagateBMCIndex(index)
}

func (this *registry) PropagateMachineIndex(index machines.MachineIndex) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, e := range this.machineindices {
		if e == index {
			return
		}
	}
	for _, c := range this.clients {
		if p, ok := c.(MachineClient); ok {
			p.PropagateMachineIndex(index)
		}
	}
	this.machineindices = append(this.machineindices, index)
}

func (this *registry) PropagateBMCIndex(index machines.BMCIndex) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, e := range this.bmcindices {
		if e == index {
			return
		}
	}
	for _, c := range this.clients {
		if p, ok := c.(BMCClient); ok {
			p.PropagateBMCIndex(index)
		}
	}
	this.bmcindices = append(this.bmcindices, index)
}
