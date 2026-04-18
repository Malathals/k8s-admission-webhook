package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func handleError(w http.ResponseWriter, message string, status int) {
	http.Error(w, message, status)
}

func parseAdmissionRequest(w http.ResponseWriter, r *http.Request) (*AdmissionReview, *Pod, bool) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(w, "Failed to read request body", http.StatusBadRequest)
		return nil, nil, false
	}

	var requestData AdmissionReview
	if err = json.Unmarshal(requestBody, &requestData); err != nil {
		handleError(w, "Failed to unmarshal admission review", http.StatusBadRequest)
		return nil, nil, false
	}

	var pod Pod
	if err = json.Unmarshal(requestData.Request.Object, &pod); err != nil {
		handleError(w, "Failed to unmarshal pod object", http.StatusBadRequest)
		return nil, nil, false
	}

	return &requestData, &pod, true
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
	requestData, pod, ok := parseAdmissionRequest(w, r)
	if !ok {
		return
	}

	patches := []map[string]interface{}{}

	if pod.Metadata.Labels == nil {
		podPatch := map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels",
			"value": map[string]string{},
		}
		patches = append(patches, podPatch)
	}

	podlabel := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/admitted-by",
		"value": "webhook",
	}
	patches = append(patches, podlabel)

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		handleError(w, "Failed to marshal JSON patch", http.StatusInternalServerError)
		return
	}

	patchType := "JSONPatch"

	response := AdmissionReview{
		APIVersion: "admission.k8s.io/v1",
		Kind:       "AdmissionReview",
		Response: &AdmissionResponse{
			UID:       requestData.Request.UID,
			Allowed:   true,
			Patch:     patchBytes,
			PatchType: &patchType,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	requestData, pod, ok := parseAdmissionRequest(w, r)
	if !ok {
		return
	}

	for _, container := range pod.Spec.Containers {
		if strings.HasSuffix(container.Image, ":latest") {
			response := AdmissionReview{
				APIVersion: "admission.k8s.io/v1",
				Kind:       "AdmissionReview",
				Response: &AdmissionResponse{
					UID:     requestData.Request.UID,
					Allowed: false,
					Result: &Status{
						Message: fmt.Sprintf("Container '%s' uses 'latest' tag, which is not allowed", container.Name),
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		if container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil && *container.SecurityContext.RunAsUser == 0 {
			response := AdmissionReview{
				APIVersion: "admission.k8s.io/v1",
				Kind:       "AdmissionReview",
				Response: &AdmissionResponse{
					UID:     requestData.Request.UID,
					Allowed: false,
					Result: &Status{
						Message: fmt.Sprintf("Container '%s' is running as root user, which is not allowed", container.Name),
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	response := AdmissionReview{
		APIVersion: "admission.k8s.io/v1",
		Kind:       "AdmissionReview",
		Response: &AdmissionResponse{
			UID:     requestData.Request.UID,
			Allowed: true,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/mutate", handleMutate)
	http.HandleFunc("/validate", handleValidate)

	fmt.Println("Starting webhook server on port 8443...")
	log.Fatal(http.ListenAndServeTLS(":8443", "/certs/tls.crt", "/certs/tls.key", nil))
}
