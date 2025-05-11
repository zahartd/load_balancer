package balancer

import (
	"log"

	"github.com/zahartd/load_balancer/internal/algorithms"
)

func CreateAlgorithm(algorithmType string) Algorithm {
	var algorithm Algorithm
	switch algorithmType {
	case "roundrobin":
		algorithm = algorithms.NewRoundRobinAlghoritm()
	default:
		log.Fatalf("Uknown algorithm type type: %s", algorithmType)
	}
	return algorithm
}
