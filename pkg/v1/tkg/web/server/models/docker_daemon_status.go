// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// DockerDaemonStatus docker daemon status
// swagger:model DockerDaemonStatus
type DockerDaemonStatus struct {

	// status
	Status bool `json:"status,omitempty"`
}

// Validate validates this docker daemon status
func (m *DockerDaemonStatus) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *DockerDaemonStatus) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *DockerDaemonStatus) UnmarshalBinary(b []byte) error {
	var res DockerDaemonStatus
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
