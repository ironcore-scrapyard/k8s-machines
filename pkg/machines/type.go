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

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"

	api "github.com/onmetal/k8s-machines/pkg/apis/machines/v1alpha1"
)

func NewMachineType(m *api.MachineType) (*MachineType, error) {
	l := len(m.Spec.MACPrefixes)
	if l == 0 {
		return nil, fmt.Errorf("prefixed required")
	}
	prefixes := make([]*MACPrefix, l, l)

	for i, p := range m.Spec.MACPrefixes {
		prefix, err := ParseMACPrefix(p)
		if err != nil {
			return nil, fmt.Errorf("invalid prefix (%d) %s: %s", i, p, err)
		}
		prefixes[i] = prefix
	}
	return &MachineType{
		Name:     resources.NewObjectName(m.Namespace, m.Name),
		prefixes: prefixes,
	}, nil
}

func ValidateMachineType(logger logger.LogContext, obj resources.Object) (*MachineType, error, error) {
	m, err := NewMachineType(obj.Data().(*api.MachineType))

	if err != nil {
		logger.Errorf("invalid machine: %s", err)
		_, err2 := resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
			m := mod.Data().(*api.MachineType)
			mod.AssureStringValue(&m.Status.State, api.STATE_INVALID)
			mod.AssureStringValue(&m.Status.Message, err.Error())
			return nil
		})
		return nil, err, err2
	}
	_, err = resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
		m := mod.Data().(*api.MachineType)
		mod.AssureStringValue(&m.Status.State, api.STATE_OK)
		mod.AssureStringValue(&m.Status.Message, "machine type ok")
		return nil
	})
	return m, nil, err
}
