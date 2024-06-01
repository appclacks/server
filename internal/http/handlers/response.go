package handlers

import "github.com/appclacks/go-client"

func NewResponse(messages ...string) client.Response {
	return client.Response{
		Messages: messages,
	}
}
