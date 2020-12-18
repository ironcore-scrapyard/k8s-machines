/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
 * This file is licensed under the Apache Software License, v. 2 except as noted
 * otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package machines

import (
	"time"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"

	"github.com/onmetal/k8s-machines/pkg/apis/machines/crds"
	api "github.com/onmetal/k8s-machines/pkg/apis/machines/v1alpha1"
)

const NAME = "dhcpleases"

const CMD_SCAN = "scan"

func init() {
	crds.AddToRegistry(apiextensions.DefaultRegistry())
}

func init() {
	controller.Configure(NAME).
		OptionsByExample("options", &Config{}).
		Reconciler(Create).
		DefaultWorkerPool(2, 0).
		MainResourceByGK(api.DHCPLEASE).
		WorkerPool("scan", 1, 30*time.Second).Commands(CMD_SCAN).
		MustRegister()
}

///////////////////////////////////////////////////////////////////////////////

func Create(controller controller.Interface) (reconcile.Interface, error) {
	cfg, _ := controller.GetOptionSource("options")
	config := cfg.(*Config)

	leases, err := NewLeaseManagement(config.LeaseFile)
	if err != nil {
		return nil, err
	}

	resc, err := controller.GetMainCluster().Resources().Get(api.DHCPLEASE)
	if err != nil {
		return nil, err
	}
	this := &reconciler{
		controller: controller,
		config:     config,
		leases:     leases,
		resource:   resc,
	}
	return this, nil
}
