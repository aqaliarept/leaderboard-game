// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// LeaderboardResponse leaderboard response
//
// swagger:model LeaderboardResponse
type LeaderboardResponse struct {

	// Timestamp of leaderboard end
	// Format: date-time
	EndsAt strfmt.DateTime `json:"ends_at,omitempty"`

	// leaderboard
	Leaderboard []*PlayerScore `json:"leaderboard"`

	// ID of the leaderboard
	LeaderboardID string `json:"leaderboard_id,omitempty"`
}

// Validate validates this leaderboard response
func (m *LeaderboardResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateEndsAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateLeaderboard(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *LeaderboardResponse) validateEndsAt(formats strfmt.Registry) error {
	if swag.IsZero(m.EndsAt) { // not required
		return nil
	}

	if err := validate.FormatOf("ends_at", "body", "date-time", m.EndsAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *LeaderboardResponse) validateLeaderboard(formats strfmt.Registry) error {
	if swag.IsZero(m.Leaderboard) { // not required
		return nil
	}

	for i := 0; i < len(m.Leaderboard); i++ {
		if swag.IsZero(m.Leaderboard[i]) { // not required
			continue
		}

		if m.Leaderboard[i] != nil {
			if err := m.Leaderboard[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("leaderboard" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("leaderboard" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this leaderboard response based on the context it is used
func (m *LeaderboardResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateLeaderboard(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *LeaderboardResponse) contextValidateLeaderboard(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Leaderboard); i++ {

		if m.Leaderboard[i] != nil {

			if swag.IsZero(m.Leaderboard[i]) { // not required
				return nil
			}

			if err := m.Leaderboard[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("leaderboard" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("leaderboard" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *LeaderboardResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *LeaderboardResponse) UnmarshalBinary(b []byte) error {
	var res LeaderboardResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
