package client

type Balancer int32

const (
	RoundRobin Balancer = iota
	RoundRand
)

var currentBalancer Balancer = RoundRobin

func SetBalancer(balancer Balancer) {
	currentBalancer = balancer
}

var maxOpenConn int64 = 16

func MaxOpenConn() int64 {
	return maxOpenConn
}

func SetMaxOpenConn(size int64) {
	maxOpenConn = size
}
