// Code generated by go-swagger; DO NOT EDIT.

package aws

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewGetAWSSubnetsParams creates a new GetAWSSubnetsParams object
// no default values defined in spec.
func NewGetAWSSubnetsParams() GetAWSSubnetsParams {

	return GetAWSSubnetsParams{}
}

// GetAWSSubnetsParams contains all the bound params for the get a w s subnets operation
// typically these are obtained from a http.Request
//
// swagger:parameters getAWSSubnets
type GetAWSSubnetsParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*VPC Id
	  Required: true
	  In: query
	*/
	VpcID string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetAWSSubnetsParams() beforehand.
func (o *GetAWSSubnetsParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qVpcID, qhkVpcID, _ := qs.GetOK("vpcId")
	if err := o.bindVpcID(qVpcID, qhkVpcID, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindVpcID binds and validates parameter VpcID from query.
func (o *GetAWSSubnetsParams) bindVpcID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("vpcId", "query")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// AllowEmptyValue: false
	if err := validate.RequiredString("vpcId", "query", raw); err != nil {
		return err
	}

	o.VpcID = raw

	return nil
}
