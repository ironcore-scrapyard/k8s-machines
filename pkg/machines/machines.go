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

package machines

import (
	"sync"
	"sync/atomic"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"k8s.io/apimachinery/pkg/labels"

	api "github.com/onmetal/k8s-machines/pkg/apis/machines/v1alpha1"
)

type Machine struct {
	Name resources.ObjectName
	*api.MachineInfoSpec
}

type Machines struct {
	initlock    sync.RWMutex
	lock        sync.RWMutex
	initialized int32
	elements    map[resources.ObjectName]*Machine
	byMACs      map[string]*Machine
	byUUIDs     map[string]*Machine
}

func NewMachines() *Machines {
	m := &Machines{
		elements: map[resources.ObjectName]*Machine{},
		byMACs:   map[string]*Machine{},
		byUUIDs:  map[string]*Machine{},
	}
	m.initlock.Lock()
	return m
}

func (this *Machines) Wait() {
	this.initlock.RLock()
	this.initlock.RUnlock()
}

func (this *Machines) Setup(logger logger.LogContext, cluster cluster.Interface) error {
	if atomic.LoadInt32(&this.initialized) != 0 {
		logger.Infof("machine cache already initialized")
		return nil
	}
	if cluster == nil {
		logger.Infof("waiting for machine cache")
		this.Wait()
		return nil
	}

	resc, err := cluster.Resources().Get(api.MACHINEINFO)
	if err != nil {
		return err
	}
	logger.Infof("setup machines")
	list, _ := resc.ListCached(labels.Everything())

	for _, l := range list {
		elem, err := this.Update(logger, l)
		if elem != nil {
			logger.Infof("found machine %s", elem.Name)
		}
		if err != nil {
			logger.Infof("errorneous machine %s: %s", l.GetName(), err)
		}
	}
	logger.Infof("machine cache setup done")
	atomic.StoreInt32(&this.initialized, 1)
	this.initlock.Unlock()
	return nil
}

func (this *Machines) Set(m *Machine) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	old := this.elements[m.Name]
	if old != nil {
		this.cleanup(old)
	}
	this.set(m)
	return nil
}

func (this *Machines) Delete(logger logger.LogContext, name resources.ObjectName) {
	this.lock.Lock()
	defer this.lock.Unlock()
	old := this.elements[name]
	if old != nil {
		this.cleanup(old)
	}
}

func (this *Machines) cleanup(m *Machine) {
	for _, n := range m.NICs {
		delete(this.byMACs, n.MAC)
	}
	delete(this.byUUIDs, m.UUID)
	delete(this.elements, m.Name)
}

func (this *Machines) set(m *Machine) {
	for _, n := range m.NICs {
		this.byMACs[n.MAC] = m
	}
	if m.UUID != "" {
		this.byUUIDs[m.UUID] = m
	}
	this.elements[m.Name] = m
}

func (this *Machines) Update(logger logger.LogContext, obj resources.Object) (*Machine, error) {
	m, err := NewMachine(obj.Data().(*api.MachineInfo))
	if err == nil {
		err = this.Set(m)
	}
	if err != nil {
		logger.Errorf("invalid machine: %s", err)
		_, err2 := resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
			m := mod.Data().(*api.MachineInfo)
			mod.AssureStringValue(&m.Status.State, api.STATE_INVALID)
			mod.AssureStringValue(&m.Status.Message, err.Error())
			return nil
		})
		return nil, err2
	}
	_, err = resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
		m := mod.Data().(*api.MachineInfo)
		mod.AssureStringValue(&m.Status.State, api.STATE_OK)
		mod.AssureStringValue(&m.Status.Message, "machine ok")
		return nil
	})
	return m, err
}

func NewMachine(m *api.MachineInfo) (*Machine, error) {
	values := m.Spec.Values.Values
	nics := m.Spec.NICs
	if values == nil {
		values = simple.Values{}
	}

	if nics == nil {
		nics = []api.NIC{}
	}
	return &Machine{
		Name:            resources.NewObjectName(m.Namespace, m.Name),
		MachineInfoSpec: &m.Spec,
	}, nil
}
