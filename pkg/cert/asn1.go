package cert

import (
	"crypto/x509/pkix"
	"encoding/asn1"
)

var OIDASN1UnstructuredName = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 2} //unstructuredName https://oidref.com/1.2.840.113549.1.9.2
var OIDASN1EmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}     //unstructuredName https://oidref.com/1.2.840.113549.1.9.1

func NewASN1UnstructuredName(val string) pkix.AttributeTypeAndValue {
	return pkix.AttributeTypeAndValue{
		Type: OIDASN1UnstructuredName,
		Value: asn1.RawValue{
			Tag:   asn1.TagIA5String,
			Bytes: []byte(val),
		},
	}
}

func NewASN1EmailAddress(val string) pkix.AttributeTypeAndValue {
	return pkix.AttributeTypeAndValue{
		Type: OIDASN1EmailAddress,
		Value: asn1.RawValue{
			Tag:   asn1.TagIA5String,
			Bytes: []byte(val),
		},
	}
}
