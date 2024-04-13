package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ApplicationStatus represents the status of the application
type ApplicationStatus struct {
	Cpu      ResourceMetrics `json:"cpu"`
	Replicas int             `json:"replicas"`
}

// ResourceMetrics represents the metrics of the application
type ResourceMetrics struct {
	HighPriority float64 `json:"highPriority"`
}

// ScalingRequest represents the request to scale the application
type ScalingRequest struct {
	Replicas int `json:"replicas"`
}

// monitorMetrics monitors the metrics of the application and scales it accordingly
func monitorMetrics(metricsHost string, metricsEndpoint string, thresholdCpu float64, maxReplicas int, minReplicas int, replicasEndpoint string) {
	metrics, err := scanMetrics(metricsHost, metricsEndpoint)

	// Handle error in monitoring metrics and skip scaling
	if err != nil {
		fmt.Printf("Error scanning metrics: %s\n", err)
		return
	}

	fmt.Printf("Metrics: %+v\n", metrics)
	if ShouldScaleUp(metrics, thresholdCpu, maxReplicas) {
		fmt.Printf("CPU is high: %f. Scaling UP to %d\n", metrics.Cpu.HighPriority, metrics.Replicas+1)
		err = scaleApplication(metricsHost, replicasEndpoint, metrics.Replicas+1)
	}
	if ShouldScaleDown(metrics, thresholdCpu, minReplicas) {
		fmt.Printf("CPU is low: %f. Scaling DOWN to %d\n", metrics.Cpu.HighPriority, metrics.Replicas-1)
		// Can add a check to not scale down unless continuous 3 times CPU usage is below threshold
		err = scaleApplication(metricsHost, replicasEndpoint, metrics.Replicas-1)
	}
	if err != nil {
		fmt.Printf("Error scaling application: %s\n", err)
	}
}

// ShouldScaleUp checks if the application should be scaled up
func ShouldScaleUp(metrics ApplicationStatus, thresholdCpu float64, maxReplicas int) bool {
	if metrics.Cpu.HighPriority > thresholdCpu && metrics.Replicas < maxReplicas {
		return true
	}
	return false
}

// ShouldScaleDown checks if the application should be scaled down
func ShouldScaleDown(metrics ApplicationStatus, thresholdCpu float64, minReplicas int) bool {
	if metrics.Cpu.HighPriority < thresholdCpu && metrics.Replicas > minReplicas {
		return true
	}
	return false
}

// scanMetrics fetches the metrics from the application
func scanMetrics(metricsHost string, metricsEndpoint string) (ApplicationStatus, error) {
	client := &http.Client{
		Timeout: time.Second * 3,
	}

	req, err := http.NewRequest("GET", metricsHost+metricsEndpoint, nil)
	if err != nil {
		return ApplicationStatus{}, fmt.Errorf("error building request for %s: %w", metricsHost+metricsEndpoint, err)
	}

	req.Header.Add("Accept", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return ApplicationStatus{}, fmt.Errorf("error fetching metrics from %s: %w", metricsHost+metricsEndpoint, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return ApplicationStatus{}, fmt.Errorf("error reading metrics from %s: %w", metricsHost+metricsEndpoint, err)
	}

	var app ApplicationStatus
	if err := json.Unmarshal(body, &app); err != nil {
		return ApplicationStatus{}, fmt.Errorf("error parsing metrics from %s: %w", metricsHost+metricsEndpoint, err)
	}

	return app, nil
}

// scaleApplication scales the application by updating the number of replicas
func scaleApplication(metricsHost string, replicasEndpoint string, replicas int) error {
	client := &http.Client{
		Timeout: time.Second * 3,
	}

	reqBody := ScalingRequest{
		Replicas: replicas,
	}
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(reqBody)

	// Call API /app/replicas to scale application
	req, err := http.NewRequest("PUT", metricsHost+replicasEndpoint, reqBodyBytes)
	if err != nil {
		return fmt.Errorf("error building request for %s: %w", metricsHost+replicasEndpoint, err)
	}

	req.Header.Add("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error scaling application from %s: %w", metricsHost+replicasEndpoint, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("received non-OK status code: %d", response.StatusCode)
	}

	return nil
}
