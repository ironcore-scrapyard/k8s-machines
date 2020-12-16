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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Struct struct{}

var _ = Describe("Mac Addresses", func() {

	Context("Parse Prefix", func() {
		It("handle single byte prefix", func() {
			p, err := ParseMACPrefix("20/4")
			Expect(err).To(BeNil())

			Expect(p.String()).To(Equal("20:00:00:00:00:00/4"))
		})
		It("handle multi byte prefix", func() {
			p, err := ParseMACPrefix("22:40:60/4")
			Expect(err).To(BeNil())

			Expect(p.String()).To(Equal("20:00:00:00:00:00/4"))
		})
	})

	Context("Prefix Containss", func() {
		mac1, _ := ParseMAC("21:22:23:24:25:16")
		mac2, _ := ParseMAC("31:22:23:24:25:16")
		mac3, _ := ParseMAC("21:80:23:24:25:16")
		mac4, _ := ParseMAC("22:80:23:24:25:16")

		It("handle single byte prefix", func() {
			p, err := ParseMACPrefix("20/4")
			Expect(err).To(BeNil())

			Expect(p.Contains(mac1)).To(BeTrue())
		})
		It("handle single byte prefix", func() {
			p, err := ParseMACPrefix("20/8")
			Expect(err).To(BeNil())

			Expect(p.Contains(mac1)).To(BeFalse())
		})

		It("handle single byte prefix", func() {
			p, err := ParseMACPrefix("20/4")
			Expect(err).To(BeNil())

			Expect(p.Contains(mac2)).To(BeFalse())
		})
		It("handle single byte prefix", func() {
			p, err := ParseMACPrefix("20/4")
			Expect(err).To(BeNil())

			Expect(p.Contains(mac4)).To(BeTrue())
		})
		It("handle single byte prefix", func() {
			p, err := ParseMACPrefix("20/8")
			Expect(err).To(BeNil())

			Expect(p.Contains(mac4)).To(BeFalse())
		})
		It("handle single byte prefix", func() {
			p, err := ParseMACPrefix("21/8")
			Expect(err).To(BeNil())

			Expect(p.Contains(mac3)).To(BeTrue())
		})

	})
})
