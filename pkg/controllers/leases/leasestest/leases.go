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

package leasestest

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onmetal/k8s-machines/pkg/controllers/leases"
)

func init() {
	leases.Register("test", NewLeaseManagement)
}

////////////////////////////////////////////////////////////////////////////////

type lease struct {
	Namespace  string      `json:"Namespace"`
	Hostname   string      `json:"Hostname"`
	IP         string      `json:"IP"`
	MAC        string      `json:"MAC"`
	LeaseTime  metav1.Time `json:"Lease"`
	ExpireTime metav1.Time `json:"Expire"`
}

type leaseManagement struct {
	lock sync.Mutex
	path string
}

var _ leases.LeaseManagement = &leaseManagement{}

func NewLeaseManagement(path string) (leases.LeaseManagement, error) {
	return &leaseManagement{path: path}, nil
}

func (this *leaseManagement) Start(c controller.Interface) error {
	// filewatcher.Configure().For("testfile").EnqueueCommand(c, leases.CMD_SCAN).StartWith(c, "leasefile")
	c.EnqueueCommand(leases.CMD_SCAN)
	return nil
}

func (this *leaseManagement) List() ([]*leases.Lease, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	list, err := this.list()
	if err != nil {
		return nil, err
	}
	result := []*leases.Lease{}
	for _, l := range list {
		result = append(result, l)
	}
	logger.Infof("found %d leases", len(result))
	return result, nil
}

func (this *leaseManagement) list() (map[string]*leases.Lease, error) {
	result := map[string]*leases.Lease{}
	list, err := ioutil.ReadDir(this.path)
	if err != nil {
		return nil, err
	}
	for _, e := range list {
		data, err := ioutil.ReadFile(filepath.Join(this.path, e.Name()))
		if err != nil {
			logger.Infof("error reading file %q: %s", e.Name(), err)
			continue
		}
		var tmp lease
		err = yaml.Unmarshal(data, &tmp)
		if err != nil {
			logger.Infof("error parsing file %q: %s", e.Name(), err)
			continue
		}
		ip := net.ParseIP(tmp.IP)
		if ip == nil {
			logger.Infof("error parsing ip %q: %s", e.Name(), tmp.IP)
			continue
		}
		mac, err := net.ParseMAC(tmp.MAC)
		if err != nil {
			logger.Infof("error parsing ip %q: %s", e.Name(), err)
			continue
		}
		logger.Infof("found lease for %s -> %s", mac, ip)
		lease := &leases.Lease{
			Namespace:  tmp.Namespace,
			Hostname:   tmp.Hostname,
			LeaseTime:  tmp.LeaseTime.Time,
			ExpireTime: tmp.ExpireTime.Time,
			IP:         ip,
			MAC:        mac,
		}
		result[e.Name()] = lease
	}
	return result, nil
}

func (this *leaseManagement) Get(mac string) *leases.Lease {
	this.lock.Lock()
	defer this.lock.Unlock()

	_, l := this.get(mac)
	return l
}

func (this *leaseManagement) get(mac string) (string, *leases.Lease) {
	list, err := this.list()
	if err == nil {
		for n, l := range list {
			if l.MAC.String() == mac {
				return n, l
			}
		}
	}
	return "", nil
}

func (this *leaseManagement) Create(l *leases.Lease) error {
	n := l.MAC.String()
	if this.Get(n) != nil {
		return fmt.Errorf("already exsists")
	}
	return this.write(n, l)
}

func (this *leaseManagement) Update(l *leases.Lease) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	n, _ := this.get(l.MAC.String())
	if n == "" {
		return fmt.Errorf("lease not found")
	}
	return this.write(n, l)
}

func (this *leaseManagement) write(n string, l *leases.Lease) error {
	data, err := yaml.Marshal(&lease{
		Hostname:   l.Hostname,
		Namespace:  l.Namespace,
		IP:         l.IP.String(),
		MAC:        l.MAC.String(),
		LeaseTime:  metav1.NewTime(l.LeaseTime),
		ExpireTime: metav1.NewTime(l.ExpireTime),
	})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(this.path, n), data, 0660)
}

func (this *leaseManagement) Delete(mac string) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	n, _ := this.get(mac)
	if n == "" {
		return nil
	}
	return os.Remove(filepath.Join(this.path, n))
}
