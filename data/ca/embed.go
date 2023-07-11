package ca

import _ "embed"

//go:embed ca.cert.pem
var CACert []byte

//go:embed ca.key.pem
var CAKey []byte
