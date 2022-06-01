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
	"fmt"
	"strings"

	"k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// storage resource quota by user.
func admitPVC(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("admitting pvc")
	pvcResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "PersistentVolumeClaim"}
	if ar.Request.Resource != pvcResource {
		err := fmt.Errorf("expect resource to be %s", pvcResource)
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	raw := ar.Request.Object.Raw
	pvc := corev1.PersistentVolumeClaim{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pvc); err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}
	// check resource quota
	reviewResponse := v1.AdmissionResponse{}
	username := ar.Request.UserInfo.Username
	//  get quota by user
	need := pvc.Spec.Resources.Requests.Storage().Value()
	total, used, _ := getQuota(username)
	switch ar.Request.Operation {
	case v1.Create:
		// 	check and update quota status
		if total-used < need {
			// deny
			reviewResponse.Allowed = false
			reviewResponse.Result = &metav1.Status{Message: strings.TrimSpace("quota limit")}
			return &reviewResponse
		}
		// 	TODO: allow and update quota status
		updateQuota(username, used+need)
		reviewResponse.Allowed = true
		return &reviewResponse
	case v1.Update:
		oldPVC := corev1.PersistentVolumeClaim{}
		deserializer = codecs.UniversalDeserializer()
		if _, _, err := deserializer.Decode(ar.Request.OldObject.Raw, nil, &oldPVC); err != nil {
			klog.Error(err)
			return toV1AdmissionResponse(err)
		}
		oldNeed := oldPVC.Spec.Resources.Requests.Storage().Value()
		if oldNeed == need {
			reviewResponse.Allowed = true
			return &reviewResponse
		}
		if oldNeed < need {
			delta := oldNeed - need
			updateQuota(username, used+delta)
			reviewResponse.Allowed = true
			return &reviewResponse
		}

		// 	other need check
		if total-used < need {
			// deny
			reviewResponse.Allowed = false
			reviewResponse.Result = &metav1.Status{Message: strings.TrimSpace("quota limit")}
			return &reviewResponse
		}
		// 	TODO: allow and update quota status
		updateQuota(username, used+need)
		reviewResponse.Allowed = true
		return &reviewResponse

	// 	check and update quota status
	case v1.Delete:
		updateQuota(username, used-need)
		reviewResponse.Allowed = true
		return &reviewResponse
	// 	update quota status
	case v1.Connect:
		// what?
	}
	reviewResponse.Allowed = true
	reviewResponse.Result = &metav1.Status{Message: strings.TrimSpace(fmt.Sprintf("invalid operation(%s)", ar.Request.Operation))}
	return &reviewResponse
}

func getQuota(username string) (total, used int64, err error) {
	_ = username
	// get quota by username
	return 100, 80, nil
}
func updateQuota(username string, quota int64) {
	_ = username
	_ = quota
	// update crd quota
}
