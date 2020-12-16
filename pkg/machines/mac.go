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
	"strconv"
	"strings"
)

type MAC = net.HardwareAddr

const EUI48_LEN = 6
const EUI64_LEN = 8

func ParseMAC(s string) (net.HardwareAddr, error) {
	return net.ParseMAC(s)
}

type MACPrefix struct {
	Address MAC
	Bits    int
}

func (this *MACPrefix) String() string {
	return fmt.Sprintf("%s/%d", this.Address, this.Bits)
}

func (this *MACPrefix) Contains(m MAC) bool {
	if this.Bits > len(m)*8 {
		return false
	}

	b := 0
	mask := byte(0xFF)
	for b < this.Bits {
		if this.Bits-b < 8 {
			mask = ^byte(0xff>>this.Bits - b)
		}
		if m[b/8]&mask != this.Address[b/8]&mask {
			return false
		}
		b += 8
	}
	return true
}

func ParseMACPrefix(s string) (*MACPrefix, error) {
	i := strings.Index(s, "/")
	if i < 0 {
		return nil, fmt.Errorf("%q is no MAC Prefix", s)
	}

	m, err := parseMACPart(s[:i])
	if err != nil {
		return nil, err
	}

	_l, err := strconv.ParseInt(s[i+1:], 10, 8)
	l := int(_l)
	if err != nil || l > 20*8 || l < 0 {
		return nil, fmt.Errorf("incalid length %q", s[i+1:])
	}
	r := 20
	if l < EUI48_LEN*8 {
		r = EUI48_LEN
	} else {
		if l < EUI64_LEN*8 {
			r = EUI64_LEN
		}
	}
	b := 0
	for b < len(m)*8 {
		if b+8 > l {
			if b >= l {
				m[b/8] = 0
			} else {
				m[b/8] = m[b/8] &^ byte(0xff>>b+8-l)
			}
		}
		b += 8
	}
	for len(m) < r {
		m = append(m, 0)
	}
	return &MACPrefix{
		Address: m,
		Bits:    int(l),
	}, nil
}

func parseMACPart(s string) (hw MAC, err error) {
	if len(s) < 2 {
		goto error
	}
	if len(s) == 2 || s[2] == ':' || s[2] == '-' {
		sep := byte(':')
		if (len(s)+1)%3 != 0 {
			goto error
		}
		n := (len(s) + 1) / 3
		if n > 20 {
			goto error
		}
		hw = make(MAC, n)
		for x, i := 0, 0; i < n; i++ {
			var ok bool
			if hw[i], ok = xtoi2(s[x:], sep); !ok {
				goto error
			}
			x += 3
		}
	} else if s[4] == '.' {
		if (len(s)+1)%5 != 0 {
			goto error
		}
		n := 2 * (len(s) + 1) / 5
		if n != 6 && n != 8 && n != 20 {
			goto error
		}
		hw = make(MAC, n)
		for x, i := 0, 0; i < n; i += 2 {
			var ok bool
			if hw[i], ok = xtoi2(s[x:x+2], 0); !ok {
				goto error
			}
			if hw[i+1], ok = xtoi2(s[x+2:], s[4]); !ok {
				goto error
			}
			x += 5
		}
	} else {
		goto error
	}
	return hw, nil

error:
	return nil, &net.AddrError{Err: "invalid MAC address", Addr: s}
}

const big = 0xFFFFFF

// Hexadecimal to integer.
// Returns number, characters consumed, success.
func xtoi(s string) (n int, i int, ok bool) {
	n = 0
	for i = 0; i < len(s); i++ {
		if '0' <= s[i] && s[i] <= '9' {
			n *= 16
			n += int(s[i] - '0')
		} else if 'a' <= s[i] && s[i] <= 'f' {
			n *= 16
			n += int(s[i]-'a') + 10
		} else if 'A' <= s[i] && s[i] <= 'F' {
			n *= 16
			n += int(s[i]-'A') + 10
		} else {
			break
		}
		if n >= big {
			return 0, i, false
		}
	}
	if i == 0 {
		return 0, i, false
	}
	return n, i, true
}

// xtoi2 converts the next two hex digits of s into a byte.
// If s is longer than 2 bytes then the third byte must be e.
// If the first two bytes of s are not hex digits or the third byte
// does not match e, false is returned.
func xtoi2(s string, e byte) (byte, bool) {
	if len(s) > 2 && s[2] != e {
		return 0, false
	}
	n, ei, ok := xtoi(s[:2])
	return byte(n), ok && ei == 2
}
