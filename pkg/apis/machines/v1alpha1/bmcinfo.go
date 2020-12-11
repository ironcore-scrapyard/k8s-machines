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
	"github.com/gardener/controller-manager-library/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BaseBoardManagementControllerInfoList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BaseBoardManagementControllerInfo `json:"items"`
}

// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=bmci,path=baseboardmanagementcontrollerinfos,singular=baseboardmanagementcontrollerinfo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name=UUID,JSONPath=".spec.uuid",type=string
// +kubebuilder:printcolumn:name=State,JSONPath=".status.state",type=string
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BaseBoardManagementControllerInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BaseBoardManagementControllerInfoSpec `json:"spec"`
	// +optional
	Status OutOfBandInfoStatus `json:"status,omitempty"`
}

type BaseBoardManagementControllerInfoSpec struct {
	// +optional
	UUID string `json:"uuid"`
	// +optional
	BMCVersion string `json:"bmcVersion,omitempty"`
	// +optional
	NIC string `json:"nic,omitempty"`
	// +optional
	IP string `json:"ip,omitempty"`
	// +optional
	MAC string `json:"mac,omitempty"`
	// +optional
	Credentials *BasicAuthCredentials `json:"credentials,omitempty"`

	// +optional
	FRUs []FieldReplacableUnit `json:"frus,omitempty"`

	// +kubebuilder:validation:XPreserveUnknownFields
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Values types.Values `json:"values,omitempty"`
}

type BasicAuthCredentials struct {
	// +optional
	Password string `json:"password,omitempty"`
	// +optional
	User string `json:"user,omitempty"`
}

type FieldReplacableUnit struct { // +optional
	ID string `json:"id"`
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	Chassis *FieldReplacableUnitInfo `json:"chassis,omitempty"`
	// +optional
	Board *FieldReplacableUnitInfo `json:"board,omitempty"`
	// +optional
	Product *FieldReplacableUnitInfo `json:"product,omitempty"`
}

type FieldReplacableUnitInfo struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Type string `json:"type,omitempty"`
	// +optional
	Serial string `json:"serial,omitempty"`
	// +optional
	Manufacturer string `json:"manufacturer,omitempty"`
	// +optional
	MfgData string `json:"mfgDate,omitempty"`
	// +optional
	PartNumber string `json:"partNumber,omitempty"`
	// +optional
	Version string `json:"version,omitempty"`
	// +optional
	AssetTag string `json:"assetTag,omitempty"`
	// +optional
	Extra []string `json:"extraxtra,omitempty"`
	// +kubebuilder:validation:XPreserveUnknownFields
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Values types.Values `json:"values,omitempty"`
}

type OutOfBandInfoStatus struct {
	// +optional
	State string `json:"state"`

	// +optional
	Message string `json:"message,omitempty"`
}
