// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// AWSAccountParams a w s account params
// swagger:model AWSAccountParams
type AWSAccountParams struct {

	// access key ID
	AccessKeyID string `json:"accessKeyID,omitempty"`

	// profile name
	ProfileName string `json:"profileName,omitempty"`

	// region
	Region string `json:"region,omitempty"`

	// secret access key
	SecretAccessKey string `json:"secretAccessKey,omitempty"`

	// session token
	SessionToken string `json:"sessionToken,omitempty"`
}

// Validate validates this a w s account params
func (m *AWSAccountParams) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *AWSAccountParams) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *AWSAccountParams) UnmarshalBinary(b []byte) error {
	var res AWSAccountParams
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
