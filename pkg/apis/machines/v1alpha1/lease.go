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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DHCPLeaseList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DHCPLease `json:"items"`
}

// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=dlease,path=dhcpleases,singular=dhcplease
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name=Hostname,JSONPath=".spec.hostname",type=string,description="hostname of machine"
// +kubebuilder:printcolumn:name=MAC,JSONPath=".spec.macAddress",type=string
// +kubebuilder:printcolumn:name=IP,JSONPath=".spec.ipAddress",type=string
// +kubebuilder:printcolumn:name=Granted,JSONPath=".spec.leaseTime",type=string,description="Time until the lease has been granted"
// +kubebuilder:printcolumn:name=Expires,JSONPath=".spec.expireTime",type=string,description="Time until the lease is valid"
// +kubebuilder:printcolumn:name=State,JSONPath=".status.state",type=string
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DHCPLease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DHCPLeaseSpec `json:"spec"`
	// +optional
	Status DHCPLeaseStatus `json:"status,omitempty"`
}

type DHCPLeaseSpec struct {
	// Machine Name
	// +optional
	Hostname string `json:"hostname"`
	// Assigned IP
	IP string `json:"ipAddress"`
	// MAC Address of requesting machine
	MAC string `json:"macAddress"`

	// Time until the lease is valid
	// +optional
	LeaseTime metav1.Time `json:"leaseTime"`
	// Time until the lease is valid
	// +optional
	ExpireTime metav1.Time `json:"expireTime"`
}

type DHCPLeaseStatus struct {
	// +optional
	State string `json:"state"`

	// +optional
	Message string `json:"message,omitempty"`
}
