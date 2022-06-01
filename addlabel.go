/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	addFirstLabelPatch string = `[
         { "op": "add", "path": "/metadata/labels", "value": {"%s": "%s"}}
     ]`
	addAdditionalLabelPatch string = `[
         { "op": "add", "path": "/metadata/labels/%s", "value": "%s" }
     ]`
	updateLabelPatch string = `[
         { "op": "replace", "path": "/metadata/labels/%s", "value": "%s" }
     ]`
)
const (
	userLabel   = "a.b.c.d"
	defaultUser = "admin"
)

// Add a label {"a.b.c.d": $username} to the object(just for pvc)
func addLabel(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("calling add-label")
	obj := struct {
		metav1.ObjectMeta `json:"metadata,omitempty"`
	}{}
	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}
	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true
	username := ar.Request.UserInfo.Username
	pt := v1.PatchTypeJSONPatch
	labelValue, hasLabel := obj.ObjectMeta.Labels[userLabel]
	switch {
	case len(obj.ObjectMeta.Labels) == 0:
		reviewResponse.Patch = []byte(fmt.Sprintf(addFirstLabelPatch, userLabel, username))
		reviewResponse.PatchType = &pt
	case !hasLabel:
		reviewResponse.Patch = []byte(fmt.Sprintf(addAdditionalLabelPatch, userLabel, username))
		reviewResponse.PatchType = &pt
	case labelValue != username:
		reviewResponse.Patch = []byte(fmt.Sprintf(updateLabelPatch, userLabel, username))
		reviewResponse.PatchType = &pt
	default:
		// already set
	}
	return &reviewResponse
}
