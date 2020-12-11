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

package machineindexer

import (
	"fmt"
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/server"
	"github.com/gardener/controller-manager-library/pkg/resources"

	"github.com/onmetal/k8s-machines/pkg/controllers"
	"github.com/onmetal/k8s-machines/pkg/machines"
)

const CONTENT_TYPE = "Content-Type"

type requesthandler struct {
	server       server.Interface
	machineindex machines.MachineIndex
	bmcindex     machines.BMCIndex
}

func (this *requesthandler) Setup() error {
	this.machineindex = controllers.GetOrCreateMachineIndex(this.server.GetEnvironment())
	this.bmcindex = controllers.GetOrCreateBMCIndex(this.server.GetEnvironment())
	this.server.Register("info", this.machineInfo)
	this.server.Register("bmc", this.bmcInfo)
	return nil
}

func (this *requesthandler) params(r *http.Request) ([]string, []string) {
	values := r.URL.Query()
	this.server.Infof("  found uuids: %v", values["uuid"])
	this.server.Infof("  found macs : %v", values["mac"])
	return values["uuid"], values["mac"]
}

func (this *requesthandler) machineInfo(w http.ResponseWriter, r *http.Request) {
	var found *machines.Machine

	this.server.Infof("query machine info: %s", r.URL.RawQuery)
	if this.machineindex == nil || !this.machineindex.IsInitialized() {
		this.server.Error("no machine index found")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	uuids, macs := this.params(r)
	for _, mac := range macs {
		m := this.machineindex.GetByMAC(mac)
		this.server.Infof("mac %s -> %v", mac, m)
		if m != nil {
			if found != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			found = m
		}
	}
	for _, uuid := range uuids {
		m := this.machineindex.GetByUUID(uuid)
		this.server.Infof("uuid %s -> %v", uuid, m)
		if m != nil {
			if found != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			found = m
		}
	}
	if found != nil {
		w.Header().Set(CONTENT_TYPE, "application/json")
		writeName(w, found.Name)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (this *requesthandler) bmcInfo(w http.ResponseWriter, r *http.Request) {
	var found *machines.BaseBoardManagementController

	this.server.Infof("query bmc info: %s", r.URL.RawQuery)
	if this.bmcindex == nil || !this.bmcindex.IsInitialized() {
		this.server.Error("no bmc index found")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	uuids, macs := this.params(r)
	for _, mac := range macs {
		m := this.bmcindex.GetByMAC(mac)
		if m != nil {
			if found != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			found = m
		}
	}
	for _, uuid := range uuids {
		m := this.bmcindex.GetByUUID(uuid)
		if m != nil {
			if found != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			found = m
		}
	}
	if found != nil {
		w.Header().Set(CONTENT_TYPE, "application/json")
		writeName(w, found.Name)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func writeName(w http.ResponseWriter, n resources.ObjectName) {
	r := fmt.Sprintf("{ \"name\": \"%s\", \"namespace\": \"%s\" }", n.Name(), n.Namespace())
	w.Write([]byte(r))
}
