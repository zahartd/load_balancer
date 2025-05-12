package balancer

import (
	"log"

	balancer_algorithms "github.com/zahartd/load_balancer/internal/balancer/algorithms"
)

func CreateAlgorithm(algorithmType string) Algorithm {
	var algorithm Algorithm
	switch algorithmType {
	case "round_robin":
		algorithm = balancer_algorithms.NewRoundRobinAlghoritm()
	default:
		log.Fatalf("Uknown algorithm type type: %s", algorithmType)
	}
	return algorithm
}
