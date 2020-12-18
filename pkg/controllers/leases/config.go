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
	"fmt"
	"os"

	"github.com/gardener/controller-manager-library/pkg/config"
)

type Config struct {
	Implementation string
	LeaseFile      string
}

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	set.AddStringOption(&this.Implementation, "lease", "", "test", "lease implementation")
	set.AddStringOption(&this.LeaseFile, "lease-file", "", "", "lease file to scan")
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (this *Config) Prepare() error {
	if this.LeaseFile == "" {
		return fmt.Errorf("lease file must be specified")
	}

	if !Exists(this.LeaseFile) {
		return fmt.Errorf("lease file %q does not exist", this.LeaseFile)
	}
	return nil
}
