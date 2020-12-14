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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/utils"

	api "github.com/onmetal/k8s-machines/pkg/apis/machines/v1alpha1"
)

const PATH_MACHINEINFO = "info"
const PATH_BMCINFO = "bmc"

type entry struct {
	name  resources.ObjectName
	uuids utils.StringSet
	macs  utils.StringSet
	last  time.Time
}

type IndexServerClient struct {
	logger   logger.LogContext
	lock     sync.RWMutex
	url      *url.URL
	maxcache int

	list  []entry
	uuids map[string]*entry
	macs  map[string]*entry
}

func NewIndexServerClient(logger logger.LogContext, url *url.URL, max int) *IndexServerClient {
	return &IndexServerClient{
		logger:   logger,
		url:      url,
		maxcache: max,
		uuids:    map[string]*entry{},
		macs:     map[string]*entry{},
	}
}

func (this *IndexServerClient) _add(name resources.ObjectName) *entry {
	var found *entry
	var oldest int
	t := time.Now()
	for i, e := range this.list {
		if e.name.Namespace() == name.Namespace() && e.name.Name() == name.Name() {
			found = &e
		}
		if !t.After(e.last) {
			oldest = i
		}
	}
	if found == nil {
		this.list = append(this.list, entry{name: name, macs: utils.StringSet{}, uuids: utils.StringSet{}})
		found = &this.list[len(this.list)-1]
	}
	if len(this.list) > this.maxcache {
		d := &this.list[oldest]
		for m := range d.macs {
			delete(this.macs, m)
		}
		for u := range d.uuids {
			delete(this.uuids, u)
		}
		this.list = append(this.list[:oldest], this.list[oldest+1:]...)
	}
	return found
}

func (this *IndexServerClient) addMAC(mac string, name resources.ObjectName) {
	found := this._add(name)
	found.macs.Add(mac)

	if this.maxcache > 0 {
		this.list = append(this.list, *found)
		found = &this.list[len(this.list)-1]
	}
	this.macs[mac] = found
}

func (this *IndexServerClient) addUUID(uuid string, name resources.ObjectName) {
	found := this._add(name)
	found.uuids.Add(uuid)

	if this.maxcache > 0 {
		this.list = append(this.list, *found)
		found = &this.list[len(this.list)-1]
	}
	this.uuids[uuid] = found
}

func (this *IndexServerClient) get(mac string, uuid string) (resources.ObjectName, error) {
	var found *entry

	this.lock.RLock()

	if mac != "" {
		found = this.macs[mac]
	}
	if found == nil && uuid != "" {
		found = this.uuids[uuid]
	}
	this.lock.RUnlock()
	if found != nil {
		return found.name, nil
	}

	url := *this.url

	q := url.Query()
	if mac != "" {
		q.Set("mac", mac)
	}
	if uuid != "" {
		q.Set("uuid", uuid)
	}
	url.RawQuery = q.Encode()

	this.logger.Infof("querying %s", url.String())
	r, err := http.Get(url.String())
	if err != nil {
		this.logger.Errorf("querying %s failed: %s", url.String(), err)
		return nil, err
	}

	data, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return nil, err
	}
	resp := &IndexResponse{}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	name := resources.NewObjectName(resp.Namespace, resp.Name)

	this.lock.Lock()
	defer this.lock.Unlock()
	if mac != "" {
		this.addMAC(mac, name)
	}
	if uuid != "" {
		this.addMAC(uuid, name)
	}
	return name, nil
}

////////////////////////////////////////////////////////////////////////////////

type indexServerIndex struct {
	access   *IndexServerClient
	resource resources.Interface
}

func NewIndexServerIndex(logger logger.LogContext, url *url.URL, res resources.Interface, max int) *indexServerIndex {
	return &indexServerIndex{
		access:   NewIndexServerClient(logger, url, max),
		resource: res,
	}
}

func (this *indexServerIndex) IsInitialized() bool {
	return true
}

////////////////////////////////////////////////////////////////////////////////

type MachineIndexServerIndex struct {
	*indexServerIndex
}

var _ MachineIndex = &MachineIndexServerIndex{}

func NewMachineIndexServerIndex(logger logger.LogContext, cluster cluster.Interface, host string, port int, max int) (MachineIndex, error) {
	f, err := NewMachineIndexServerIndexCreator(logger, cluster, host, port, max)
	if err != nil {
		return nil, err
	}
	return f(), nil
}

func NewMachineIndexServerIndexCreator(logger logger.LogContext, cluster cluster.Interface, host string, port int, max int) (func() MachineIndex, error) {
	var url url.URL
	url.Scheme = "http"
	url.Host = fmt.Sprintf("%s:%d", host, port)
	url.Path = "/"+PATH_MACHINEINFO

	res, err := cluster.Resources().Get(api.MACHINEINFO)
	if err != nil {
		return nil, err
	}
	return func() MachineIndex {
		return &MachineIndexServerIndex{
			NewIndexServerIndex(logger, &url, res, max),
		}
	}, nil
}

func (this *MachineIndexServerIndex) GetByUUID(uuid string) *Machine {
	n, _ := this.access.get("", uuid)
	if n == nil {
		return nil
	}
	return this.GetByName(n)
}

func (this *MachineIndexServerIndex) GetByMAC(mac string) *Machine {
	n, _ := this.access.get(mac, "")
	if n == nil {
		return nil
	}
	return this.GetByName(n)
}

func (this *MachineIndexServerIndex) GetByName(name resources.ObjectName) *Machine {
	o, _ := this.resource.Get(name)
	m, _ := NewMachine(o.Data().(*api.MachineInfo))
	return m
}

////////////////////////////////////////////////////////////////////////////////

type BMCIndexServerIndex struct {
	*indexServerIndex
}

var _ BMCIndex = &BMCIndexServerIndex{}

func NewBMCIndexServerIndex(logger logger.LogContext, cluster cluster.Interface, host string, port int, max int) (BMCIndex, error) {
	f, err := NewBMCIndexServerIndexCreator(logger, cluster, host, port, max)
	if err != nil {
		return nil, err
	}
	return f(), nil
}

func NewBMCIndexServerIndexCreator(logger logger.LogContext, cluster cluster.Interface, host string, port int, max int) (func() BMCIndex, error) {
	var url url.URL
	url.Scheme = "http"
	url.Host = fmt.Sprintf("%s:%d", host, port)
	url.Path = "/"+PATH_BMCINFO

	res, err := cluster.Resources().Get(api.BASEBOARDMANAGEMENTCONTROLLERINFO)
	if err != nil {
		return nil, err
	}
	return func() BMCIndex {
		return &BMCIndexServerIndex{
			NewIndexServerIndex(logger, &url, res, max),
		}
	}, nil
}

func (this *BMCIndexServerIndex) GetByUUID(uuid string) *BaseBoardManagementController {
	n, _ := this.access.get("", uuid)
	if n == nil {
		return nil
	}
	return this.GetByName(n)
}

func (this *BMCIndexServerIndex) GetByMAC(mac string) *BaseBoardManagementController {
	n, _ := this.access.get(mac, "uuid")
	if n == nil {
		return nil
	}
	return this.GetByName(n)
}

func (this *BMCIndexServerIndex) GetByName(name resources.ObjectName) *BaseBoardManagementController {
	o, _ := this.resource.Get(name)
	m, _ := NewBaseBoardManagementController(o.Data().(*api.BaseBoardManagementControllerInfo))
	return m
}
