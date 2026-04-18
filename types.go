package main

import "encoding/json"

type AdmissionReview struct {
	APIVersion string             `json:"apiVersion"`
	Kind       string             `json:"kind"`
	Request    *AdmissionRequest  `json:"request,omitempty"`
	Response   *AdmissionResponse `json:"response,omitempty"`
}

type AdmissionRequest struct {
	UID    string          `json:"uid"`
	Object json.RawMessage `json:"object"`
}

type AdmissionResponse struct {
	UID       string  `json:"uid"`
	Allowed   bool    `json:"allowed"`
	Result    *Status `json:"status,omitempty"`
	Patch     []byte  `json:"patch,omitempty"`
	PatchType *string `json:"patchType,omitempty"`
}

type Status struct {
	Message string `json:"message"`
}

type Pod struct {
	Metadata PodMetadata `json:"metadata"`
	Spec     PodSpec     `json:"spec"`
}

type PodMetadata struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type PodSpec struct {
	Containers []Container `json:"containers"`
}

type Container struct {
	Name            string               `json:"name"`
	Image           string               `json:"image"`
	SecurityContext *SecurityContext     `json:"securityContext,omitempty"`
	Resources       ResourceRequirements `json:"resources,omitempty"`
}

type SecurityContext struct {
	RunAsNonRoot *bool  `json:"runAsNonRoot,omitempty"`
	RunAsUser    *int64 `json:"runAsUser,omitempty"`
}

type ResourceRequirements struct {
	Limits map[string]string `json:"limits,omitempty"`
}
