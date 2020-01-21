module github.com/gildas/go-box

go 1.13

require (
	github.com/chakrit/go-bunyan v0.0.0-20140303180041-5a9b5e7b1765 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gildas/go-core v0.4.1
	github.com/gildas/go-errors v0.0.2
	github.com/gildas/go-logger v1.3.3
	github.com/gildas/go-request v0.2.3
	github.com/google/uuid v1.1.1
	github.com/stretchr/testify v1.4.0
	github.com/youmark/pkcs8 v0.0.0-20191102193632-94c173a94d60
)

replace github.com/gildas/go-errors => ../go-errors
