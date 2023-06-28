package ca

import _ "embed"

//go:embed ca.cert.pem
var CACert []byte

//go:embed ca.key2.pem
var CAKey []byte
