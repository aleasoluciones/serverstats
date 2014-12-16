package main

import (
	"fmt"
	"github.com/aleasoluciones/serverstats"
)

func main() {
	serverStats := serverstats.NewServerStats()

	for metric := range serverStats.Metrics {
		fmt.Println(fmt.Sprintf("%-15s", metric.Name), metric.Value, metric.Unit)
	}

}