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

package bmcinfo

import (
	"net/http"

	"github.com/onmetal/k8s-machines/pkg/controllers"
	"github.com/onmetal/k8s-machines/pkg/machines"

	"github.com/onmetal/k8s-machines/pkg/servers/machineindexer"

	// register reguired controllers
	_ "github.com/onmetal/k8s-machines/pkg/controllers/bmc"
)

func init() {
	machineindexer.RegisterIndex(New)
}

type indexer struct {
	server machineindexer.IndexServer
	index  machines.BMCIndex
}

func New(server machineindexer.IndexServer) (machineindexer.Interface, error) {
	return &indexer{server: server}, nil
}

func (this *indexer) Setup() error {
	this.index = controllers.GetOrCreateBMCIndex(this.server.GetEnvironment(), func() machines.BMCIndex { return machines.NewBMCFullIndexer() })
	this.server.Register(machines.PATH_BMCINFO, this.handler)
	return nil
}

func (this *indexer) handler(w http.ResponseWriter, r *http.Request) {
	var found *machines.BaseBoardManagementController

	this.server.Infof("query bmc info: %s", r.URL.RawQuery)
	if this.index == nil || !this.index.IsInitialized() {
		this.server.Error("no bmc index found")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	uuids, macs := this.server.MachineIds(r)
	for _, mac := range macs {
		m := this.index.GetByMAC(mac)
		if m != nil {
			if found != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			found = m
		}
	}
	for _, uuid := range uuids {
		m := this.index.GetByUUID(uuid)
		if m != nil {
			if found != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			found = m
		}
	}
	if found != nil {
		w.Header().Set(machineindexer.CONTENT_TYPE, "application/json")
		this.server.ObjectResponse(w, found.Name)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
