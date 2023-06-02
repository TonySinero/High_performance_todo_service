// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type TodoElastic struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type TodoInput struct {
	Title     string `json:"title"`
	Completed *bool  `json:"completed,omitempty"`
}

type TodoInputID struct {
	ID        string  `json:"id"`
	Title     *string `json:"title,omitempty"`
	Completed *bool   `json:"completed,omitempty"`
}