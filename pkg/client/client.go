package client

import (
	cl "github.com/mpmlj/clarifai-client-go"
)

// CreateSession creates a Session from Client Auth creds
func CreateSession(apikey string) *cl.Session {
	var sess *cl.Session

	if apikey != "" {
		sess = cl.NewApp(apikey)
	}
	return sess
}
