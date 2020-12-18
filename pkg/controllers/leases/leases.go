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
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Lease struct {
	Namespace  string
	Hostname   string
	IP         net.IP
	MAC        net.HardwareAddr
	LeaseTime  time.Time
	ExpireTime time.Time
}

func (this *Lease) Equal(l *Lease) bool {
	if this.Namespace != l.Namespace {
		return false
	}
	if this.Hostname != l.Hostname {
		return false
	}
	if !this.IP.Equal(l.IP) {
		return false
	}
	if this.MAC.String() != l.MAC.String() {
		return false
	}
	if this.LeaseTime.String() != l.LeaseTime.String() {
		return false
	}
	if this.ExpireTime.String() != l.ExpireTime.String() {
		return false
	}
	return true
}

type LeaseManagement interface {
	List() ([]*Lease, error)
	Get(mac string) *Lease
	Create(*Lease) error
	Update(*Lease) error
	Delete(mac string) error
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

var _ LeaseManagement = &leaseManagement{}

func NewLeaseManagement(path string) (LeaseManagement, error) {
	return &leaseManagement{path: path}, nil
}

func (this *leaseManagement) List() ([]*Lease, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	list, err := this.list()
	if err != nil {
		return nil, err
	}
	result := []*Lease{}
	for _, l := range list {
		result = append(result, l)
	}
	logger.Infof("found %d leases", len(result))
	return result, nil
}

func (this *leaseManagement) list() (map[string]*Lease, error) {
	result := map[string]*Lease{}
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
		lease := &Lease{
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

func (this *leaseManagement) Get(mac string) *Lease {
	this.lock.Lock()
	defer this.lock.Unlock()

	_, l := this.get(mac)
	return l
}

func (this *leaseManagement) get(mac string) (string, *Lease) {
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

func (this *leaseManagement) Create(l *Lease) error {
	n := l.MAC.String()
	if this.Get(n) != nil {
		return fmt.Errorf("aleeady exsists")
	}
	return this.write(n, l)
}

func (this *leaseManagement) Update(l *Lease) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	n, _ := this.get(l.MAC.String())
	if n == "" {
		return fmt.Errorf("lease not found")
	}
	return this.write(n, l)
}

func (this *leaseManagement) write(n string, l *Lease) error {
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
