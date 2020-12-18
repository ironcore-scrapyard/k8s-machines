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

package leases

import (
	"net"
	"sync"
	"time"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
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
	Start(controller.Interface) error

	List() ([]*Lease, error)
	Get(mac string) *Lease
	Create(*Lease) error
	Update(*Lease) error
	Delete(mac string) error
}

var lock sync.Mutex
var selected = map[string]func(path string) (LeaseManagement, error){}

func Register(name string, mgmt func(path string) (LeaseManagement, error)) {
	lock.Lock()
	defer lock.Unlock()
	selected[name] = mgmt
}
