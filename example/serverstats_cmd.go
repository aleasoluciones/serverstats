package main

import (
	"fmt"
	"github.com/aleasoluciones/serverstats"
)

func main() {
	serverStats := serverstats.NewServerStats(serverstats.DefaultPeriodes)

	for metric := range serverStats.Metrics {
		fmt.Println(metric.Timestamp, fmt.Sprintf("%-15s", metric.Name), metric.Value, metric.Unit)
	}

}
