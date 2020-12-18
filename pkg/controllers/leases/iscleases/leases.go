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

package iscleases

import (
	"fmt"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"

	"github.com/onmetal/k8s-machines/pkg/controllers/leases"
	"github.com/onmetal/k8s-machines/pkg/filewatcher"
)

func init() {
	leases.Register("isclease", NewLeaseManagement)
}

////////////////////////////////////////////////////////////////////////////////

type leaseManagement struct {
	lock sync.Mutex
	path string
}

var _ leases.LeaseManagement = &leaseManagement{}

func NewLeaseManagement(path string) (leases.LeaseManagement, error) {
	return &leaseManagement{path: path}, nil
}

func (this *leaseManagement) Start(c controller.Interface) error {
	filewatcher.Configure().For(this.path).EnqueueCommand(c, leases.CMD_SCAN).StartWith(c, "leasefile")
	return nil
}

func (this *leaseManagement) List() ([]*leases.Lease, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	return nil, fmt.Errorf("not implemented yet")
}

func (this *leaseManagement) Get(mac string) *leases.Lease {
	this.lock.Lock()
	defer this.lock.Unlock()

	return nil
}

func (this *leaseManagement) Create(l *leases.Lease) error {
	n := l.MAC.String()
	if this.Get(n) != nil {
		return fmt.Errorf("already exsists")
	}
	return fmt.Errorf("not implemented yet")
}

func (this *leaseManagement) Update(l *leases.Lease) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	return fmt.Errorf("not implemented yet")
}

func (this *leaseManagement) Delete(mac string) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	return fmt.Errorf("not implemented yet")
}
