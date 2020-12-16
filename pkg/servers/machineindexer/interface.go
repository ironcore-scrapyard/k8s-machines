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
	"net/http"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/module/handler"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/server"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

type Interface = handler.SetupInterface

type IndexServer interface {
	server.Interface
	MachineIds(r *http.Request) ([]string, []string)
	ObjectResponse(w http.ResponseWriter, n resources.ObjectName)
}

type IndexHandlerType func(IndexServer) (Interface, error)

type Registry interface {
	RegisterIndex(t IndexHandlerType)
}

type registry struct {
	lock     sync.Mutex
	handlers []IndexHandlerType
}

func NewRegistry() *registry {
	return &registry{}
}

func (this *registry) RegisterIndex(t IndexHandlerType) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.handlers = append(this.handlers, t)
}

var defaultRegistry = NewRegistry()

func RegisterIndex(t IndexHandlerType) {
	defaultRegistry.RegisterIndex(t)
}
