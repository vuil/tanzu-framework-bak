// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// VSphereAvailabilityZone v sphere availability zone
// swagger:model VSphereAvailabilityZone
type VSphereAvailabilityZone struct {

	// moid
	Moid string `json:"moid,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this v sphere availability zone
func (m *VSphereAvailabilityZone) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VSphereAvailabilityZone) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VSphereAvailabilityZone) UnmarshalBinary(b []byte) error {
	var res VSphereAvailabilityZone
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
