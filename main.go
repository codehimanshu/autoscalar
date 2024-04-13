package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	// Define command line flags
	scanInterval := flag.Int("scanInterval", 5, "interval in seconds to scan resource metrics")
	thresholdCpu := flag.Float64("thresholdCpu", 0.80, "average CPU to maintain")
	minReplicas := flag.Int("minReplicas", 3, "minimum number of replicas")
	maxReplicas := flag.Int("maxReplicas", 50, "maximum number of replicas")
	metricsHost := flag.String("metricsHost", "http://localhost:8123", "Hostname to fetch application metrics")
	metricsEndpoint := flag.String("metricsEndpoint", "/app/status", "API endpoint to fetch application metrics")
	replicasEndpoint := flag.String("replicasEndpoint", "/app/replicas", "API endpoint to update application replica")
	flag.Parse()

	// Print the configuration
	fmt.Println("Starting AutoScalar")
	fmt.Printf("Resource Scan Interval for AutoScalar is %d seconds\n", *scanInterval)
	fmt.Printf("Threshold CPU for Application is %f \n", *thresholdCpu)

	// Start monitoring the metrics
	for {
		monitorMetrics(*metricsHost, *metricsEndpoint, *thresholdCpu, *maxReplicas, *minReplicas, *replicasEndpoint)
		time.Sleep(time.Duration(*scanInterval) * time.Second)
	}
}
