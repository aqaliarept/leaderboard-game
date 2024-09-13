// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// PlayerScore player score
//
// swagger:model PlayerScore
type PlayerScore struct {

	// ID of the player
	// Required: true
	PlayerID *string `json:"player_id"`

	// Player's score
	// Required: true
	Score *int64 `json:"score"`
}

// Validate validates this player score
func (m *PlayerScore) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validatePlayerID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateScore(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *PlayerScore) validatePlayerID(formats strfmt.Registry) error {

	if err := validate.Required("player_id", "body", m.PlayerID); err != nil {
		return err
	}

	return nil
}

func (m *PlayerScore) validateScore(formats strfmt.Registry) error {

	if err := validate.Required("score", "body", m.Score); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this player score based on context it is used
func (m *PlayerScore) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *PlayerScore) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *PlayerScore) UnmarshalBinary(b []byte) error {
	var res PlayerScore
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
