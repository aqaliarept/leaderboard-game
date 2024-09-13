// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/Aqaliarept/leaderboard-game/generated/server/models"
)

// GetLeaderboardOKCode is the HTTP code returned for type GetLeaderboardOK
const GetLeaderboardOKCode int = 200

/*
GetLeaderboardOK Successful response

swagger:response getLeaderboardOK
*/
type GetLeaderboardOK struct {

	/*
	  In: Body
	*/
	Payload *models.LeaderboardResponse `json:"body,omitempty"`
}

// NewGetLeaderboardOK creates GetLeaderboardOK with default headers values
func NewGetLeaderboardOK() *GetLeaderboardOK {

	return &GetLeaderboardOK{}
}

// WithPayload adds the payload to the get leaderboard o k response
func (o *GetLeaderboardOK) WithPayload(payload *models.LeaderboardResponse) *GetLeaderboardOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get leaderboard o k response
func (o *GetLeaderboardOK) SetPayload(payload *models.LeaderboardResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetLeaderboardOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetLeaderboardNotFoundCode is the HTTP code returned for type GetLeaderboardNotFound
const GetLeaderboardNotFoundCode int = 404

/*
GetLeaderboardNotFound Leaderboard not found

swagger:response getLeaderboardNotFound
*/
type GetLeaderboardNotFound struct {
}

// NewGetLeaderboardNotFound creates GetLeaderboardNotFound with default headers values
func NewGetLeaderboardNotFound() *GetLeaderboardNotFound {

	return &GetLeaderboardNotFound{}
}

// WriteResponse to the client
func (o *GetLeaderboardNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}
