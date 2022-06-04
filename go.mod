module github.com/gildas/go-box

go 1.16

require (
	github.com/gildas/go-core v0.4.10
	github.com/gildas/go-errors v0.3.2
	github.com/gildas/go-logger v1.5.5
	github.com/gildas/go-request v0.7.9
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/joho/godotenv v1.4.0
	github.com/stretchr/testify v1.7.1
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
)

replace github.com/gildas/go-request => ../go-request
