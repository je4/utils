module github.com/je4/utils/v2

go 1.16

replace github.com/je4/utils/v2 => ./

require (
	github.com/blend/go-sdk v1.20220411.3
	github.com/go-sql-driver/mysql v1.7.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gosuri/uilive v0.0.4 // indirect
	github.com/gosuri/uiprogress v0.0.1
	github.com/machinebox/progress v0.2.0
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.13.5
	golang.org/x/crypto v0.3.0
	golang.org/x/sys v0.3.0 // indirect
)
