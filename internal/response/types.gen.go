// Package response provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.1.0 DO NOT EDIT.
package response

// Response defines model for Response.
type Response struct {
	Data       interface{} `json:"data,omitempty"`
	Messages   []string    `json:"messages"`
	Successful bool        `json:"successful"`
}
