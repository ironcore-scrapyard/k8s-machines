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
	"net"
	"strings"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	api "github.com/onmetal/k8s-machines/pkg/apis/machines/v1alpha1"
)

type reconciler struct {
	reconcile.DefaultReconciler

	controller controller.Interface
	config     *Config
	leases     LeaseManagement
	resource   resources.Interface
}

var _ reconcile.Interface = &reconciler{}

func (this *reconciler) Start() error {
	this.controller.EnqueueCommand(CMD_SCAN)
	return nil
}

///////////////////////////////////////////////////////////////////////////////

func (this *reconciler) Config() *Config {
	return this.config
}

func (this *reconciler) Command(logger logger.LogContext, cmd string) reconcile.Status {
	logger.Infof("scan lease file")

	return reconcile.DelayOnError(logger, this.Update())
}

func (this *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	var err error

	l := obj.Data().(*api.DHCPLease)
	old := this.leases.Get(l.Spec.MAC)
	if old != nil {
		logger.Infof("reconcile existing lease for mac %s", l.Spec.MAC)
	} else {
		logger.Infof("reconcile new lease for mac %s", l.Spec.MAC)
	}

	if !this.controller.HasFinalizer(obj) {
		logger.Infof("found new lease object for %s", l.Spec.MAC)
		lease, err := this.createLeaseFor(l)
		if err != nil {
			return this.setState(logger, obj, api.STATE_INVALID, err.Error())
		}
		if old == nil {
			// manually created object
			logger.Infof("creating lease for %s", l.Spec.MAC)
			err = this.leases.Create(lease)
			if err != nil {
				return this.setState(logger, obj, api.STATE_INVALID, err.Error())
			}
		}
		err = this.controller.SetFinalizer(obj)
		if err != nil {
			return reconcile.Delay(logger, err)
		}
		if old == nil {
			return this.setState(logger, obj, api.STATE_OK, "")
		}
	}

	if old == nil {
		logger.Infof("lease not found -> trigger deletion")
		return reconcile.DelayOnError(logger, obj.Delete())
	}
	if this.updateObject(l, old) {
		_, err := resources.Modify(obj, func(mod *resources.ModificationState) error {
			spec := &mod.Data().(*api.DHCPLease).Spec
			mod.AssureStringValue(&spec.IP, old.IP.String())
			mod.AssureStringValue(&spec.Hostname, old.Hostname)
			mod.AssureTimeValue(&spec.LeaseTime, metav1.NewTime(old.LeaseTime))
			mod.AssureTimeValue(&spec.ExpireTime, metav1.NewTime(old.ExpireTime))
			return nil
		})
		return reconcile.DelayOnError(logger, err)
	} else {
		lease, err := this.createLeaseFor(l)
		if err != nil {
			return this.setState(logger, obj, api.STATE_INVALID, err.Error())
		}
		if !lease.Equal(old) {
			logger.Infof("update lease")
			err = this.leases.Update(lease)
			if err != nil {
				return this.setState(logger, obj, api.STATE_INVALID, err.Error())
			}
		}
	}
	if err == nil {
		return this.setState(logger, obj, api.STATE_OK, "")
	}
	return reconcile.Succeeded(logger)
}

func (this *reconciler) Delete(logger logger.LogContext, obj resources.Object) reconcile.Status {
	logger.Infof("delete")
	l := obj.Data().(*api.DHCPLease)
	err := this.leases.Delete(l.Spec.MAC)
	if err != nil {
		return reconcile.Delay(logger, err)
	}

	return reconcile.DelayOnError(logger, this.controller.RemoveFinalizer(obj))
}

////////////////////////////////////////////////////////////////////////////////

func (this *reconciler) setState(logger logger.LogContext, obj resources.Object, state string, msg string) reconcile.Status {
	return reconcile.UpdateStandardObjectStatus(logger, obj, state, msg)
}

func (this *reconciler) updateObject(o *api.DHCPLease, l *Lease) bool {
	if l.LeaseTime.After(o.Spec.LeaseTime.Time) {
		return true
	}
	if l.Hostname != o.Spec.Hostname {
		return true
	}
	if l.IP.String() != o.Spec.IP {
		return true
	}
	if l.MAC.String() != o.Spec.MAC {
		return true
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////

func (this *reconciler) Update() error {
	list, err := this.leases.List()
	if err != nil {
		return err
	}

	cur, err := this.resource.ListCached(labels.Everything())
	if err != nil {
		return err
	}

	index := map[string]*Lease{}

	for _, l := range list {
		key := l.MAC.String()
		index[key] = l
	}

	oldindex := map[string]resources.Object{}
	for _, e := range cur {
		l := e.Data().(*api.DHCPLease)
		key := l.Spec.MAC

		new := index[key]
		oldindex[key] = e
		if new == nil {
			this.controller.Enqueue(e)
		} else {
			mod := false
			if l.Spec.ExpireTime.Time.Unix() != new.ExpireTime.Unix() {
				mod = true
			}
			if l.Spec.LeaseTime.Time.Unix() != new.LeaseTime.Unix() {
				mod = true
			}
			if l.Spec.IP != new.IP.String() {
				mod = true
			}
			if l.Spec.Hostname != new.Hostname {
				mod = true
			}
			if mod {
				this.controller.Enqueue(e)
			}
		}
	}
	for key, l := range index {
		old := oldindex[key]
		if old == nil {
			err2 := this.Create(l)
			if err2 != nil {
				err = err2
			}
		}
	}
	return err
}

func (this *reconciler) Create(l *Lease) error {
	new := &api.DHCPLease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ReplaceAll(l.MAC.String(), ":", "-"),
			Namespace: l.Namespace,
		},
		Spec: api.DHCPLeaseSpec{
			LeaseTime:  metav1.NewTime(l.LeaseTime),
			ExpireTime: metav1.NewTime(l.ExpireTime),
			MAC:        l.MAC.String(),
			IP:         l.IP.String(),
			Hostname:   l.Hostname,
		},
	}
	o, err := this.resource.Wrap(new)
	new.Finalizers = []string{this.controller.FinalizerHandler().FinalizerName(o)}

	o, err = this.resource.Create(new)
	return err
}

func (this *reconciler) createLeaseFor(obj *api.DHCPLease) (*Lease, error) {
	if obj.Spec.Hostname == "" {
		return nil, fmt.Errorf("hostname missing")
	}
	if obj.Spec.MAC == "" {
		return nil, fmt.Errorf("mac missing")
	}
	mac, err := net.ParseMAC(obj.Spec.MAC)
	if err != nil {
		return nil, fmt.Errorf("invalid mac address %q: %s", obj.Spec.MAC, err)
	}
	if obj.Spec.IP == "" {
		return nil, fmt.Errorf("IP missing")
	}
	ip := net.ParseIP(obj.Spec.IP)
	if ip == nil {
		return nil, fmt.Errorf("invalid ip address %q", obj.Spec.IP)
	}
	if obj.Spec.LeaseTime.Time.IsZero() {
		return nil, fmt.Errorf("lease time missing")
	}
	if obj.Spec.ExpireTime.Time.IsZero() {
		return nil, fmt.Errorf("expire time missing")
	}
	return &Lease{
		Hostname:   obj.Spec.Hostname,
		Namespace:  obj.Namespace,
		MAC:        mac,
		IP:         ip,
		LeaseTime:  obj.Spec.LeaseTime.Time,
		ExpireTime: obj.Spec.ExpireTime.Time,
	}, nil
}
