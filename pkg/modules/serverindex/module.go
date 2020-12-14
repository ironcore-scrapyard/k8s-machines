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

package serverindex

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/module"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/module/handler"

	"github.com/onmetal/k8s-machines/pkg/controllers"
	"github.com/onmetal/k8s-machines/pkg/machines"
)

const NAME = "machineserverindex"

func init() {
	module.Configure(NAME).
		OptionsByExample("options", &Config{}).
		RegisterHandler("machineinfos", MachineInfos).
		RegisterHandler("bmcinfos", BMCInfos).
		ActivateExplicitly().
		MustRegister()
}

func MachineInfos(mod module.Interface) (handler.Interface, error) {
	opts, err := mod.GetOptionSource("options")
	if err != nil {
		return nil, err
	}
	cfg := opts.(*Config)

	mod.Infof("  using server %s:%d", cfg.Host, cfg.Port)
	mod.Infof("  using cache size %d", cfg.MaxCache)
	creator, err := machines.NewMachineIndexServerIndexCreator(mod, mod.GetMainCluster(), cfg.Host, cfg.Port, cfg.MaxCache)
	if err != nil {
		return nil, err
	}
	controllers.GetOrCreateMachineIndex(mod.GetEnvironment(), creator)
	return &Handler{}, nil
}

func BMCInfos(mod module.Interface) (handler.Interface, error) {
	opts, err := mod.GetOptionSource("options")
	if err != nil {
		return nil, err
	}
	cfg := opts.(*Config)

	creator, err := machines.NewBMCIndexServerIndexCreator(mod, mod.GetMainCluster(), cfg.Host, cfg.Port, cfg.MaxCache)
	if err != nil {
		return nil, err
	}
	controllers.GetOrCreateBMCIndex(mod.GetEnvironment(), creator)
	return &Handler{}, nil
}

type Handler struct {
}
