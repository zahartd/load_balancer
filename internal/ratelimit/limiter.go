package ratelimit

type Algorithm interface {
	Allow() bool
}
