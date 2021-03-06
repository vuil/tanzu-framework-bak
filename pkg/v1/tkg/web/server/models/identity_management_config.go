// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// IdentityManagementConfig identity management config
// swagger:model IdentityManagementConfig
type IdentityManagementConfig struct {

	// idm type
	// Required: true
	// Enum: [oidc ldap none]
	IdmType *string `json:"idm_type"`

	// ldap bind dn
	LdapBindDn string `json:"ldap_bind_dn,omitempty"`

	// ldap bind password
	LdapBindPassword string `json:"ldap_bind_password,omitempty"`

	// ldap group search base dn
	LdapGroupSearchBaseDn string `json:"ldap_group_search_base_dn,omitempty"`

	// ldap group search filter
	LdapGroupSearchFilter string `json:"ldap_group_search_filter,omitempty"`

	// ldap group search group attr
	LdapGroupSearchGroupAttr string `json:"ldap_group_search_group_attr,omitempty"`

	// ldap group search name attr
	LdapGroupSearchNameAttr string `json:"ldap_group_search_name_attr,omitempty"`

	// ldap group search user attr
	LdapGroupSearchUserAttr string `json:"ldap_group_search_user_attr,omitempty"`

	// ldap root ca
	LdapRootCa string `json:"ldap_root_ca,omitempty"`

	// ldap url
	LdapURL string `json:"ldap_url,omitempty"`

	// ldap user search base dn
	LdapUserSearchBaseDn string `json:"ldap_user_search_base_dn,omitempty"`

	// ldap user search email attr
	LdapUserSearchEmailAttr string `json:"ldap_user_search_email_attr,omitempty"`

	// ldap user search filter
	LdapUserSearchFilter string `json:"ldap_user_search_filter,omitempty"`

	// ldap user search id attr
	LdapUserSearchIDAttr string `json:"ldap_user_search_id_attr,omitempty"`

	// ldap user search name attr
	LdapUserSearchNameAttr string `json:"ldap_user_search_name_attr,omitempty"`

	// ldap user search username
	LdapUserSearchUsername string `json:"ldap_user_search_username,omitempty"`

	// oidc claim mappings
	OidcClaimMappings map[string]string `json:"oidc_claim_mappings,omitempty"`

	// oidc client id
	OidcClientID string `json:"oidc_client_id,omitempty"`

	// oidc client secret
	OidcClientSecret string `json:"oidc_client_secret,omitempty"`

	// oidc provider name
	OidcProviderName string `json:"oidc_provider_name,omitempty"`

	// oidc provider url
	// Format: uri
	OidcProviderURL strfmt.URI `json:"oidc_provider_url,omitempty"`

	// oidc scope
	OidcScope string `json:"oidc_scope,omitempty"`

	// oidc skip verify cert
	OidcSkipVerifyCert bool `json:"oidc_skip_verify_cert,omitempty"`
}

func (m *IdentityManagementConfig) UnmarshalJSON(b []byte) error {
	type IdentityManagementConfigAlias IdentityManagementConfig
	var t IdentityManagementConfigAlias
	if err := json.Unmarshal([]byte("{\"idm_type\":\"oidc\",\"ldap_group_search_name_attr\":\"cn\",\"ldap_group_search_user_attr\":\"DN\",\"ldap_user_search_email_attr\":\"userPrincipalName\",\"ldap_user_search_id_attr\":\"DN\",\"ldap_user_search_username\":\"userPrincipalName\",\"oidc_skip_verify_cert\":false}"), &t); err != nil {
		return err
	}
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	*m = IdentityManagementConfig(t)
	return nil
}

// Validate validates this identity management config
func (m *IdentityManagementConfig) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateIdmType(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOidcProviderURL(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var identityManagementConfigTypeIdmTypePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["oidc","ldap","none"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		identityManagementConfigTypeIdmTypePropEnum = append(identityManagementConfigTypeIdmTypePropEnum, v)
	}
}

const (

	// IdentityManagementConfigIdmTypeOidc captures enum value "oidc"
	IdentityManagementConfigIdmTypeOidc string = "oidc"

	// IdentityManagementConfigIdmTypeLdap captures enum value "ldap"
	IdentityManagementConfigIdmTypeLdap string = "ldap"

	// IdentityManagementConfigIdmTypeNone captures enum value "none"
	IdentityManagementConfigIdmTypeNone string = "none"
)

// prop value enum
func (m *IdentityManagementConfig) validateIdmTypeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, identityManagementConfigTypeIdmTypePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *IdentityManagementConfig) validateIdmType(formats strfmt.Registry) error {

	if err := validate.Required("idm_type", "body", m.IdmType); err != nil {
		return err
	}

	// value enum
	if err := m.validateIdmTypeEnum("idm_type", "body", *m.IdmType); err != nil {
		return err
	}

	return nil
}

func (m *IdentityManagementConfig) validateOidcProviderURL(formats strfmt.Registry) error {

	if swag.IsZero(m.OidcProviderURL) { // not required
		return nil
	}

	if err := validate.FormatOf("oidc_provider_url", "body", "uri", m.OidcProviderURL.String(), formats); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *IdentityManagementConfig) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *IdentityManagementConfig) UnmarshalBinary(b []byte) error {
	var res IdentityManagementConfig
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
