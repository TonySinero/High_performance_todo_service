package graph

//go:generate go run github.com/99designs/gqlgen generate

import "newFeatures/service"

type Resolver struct {
	Serv *service.Service
}
