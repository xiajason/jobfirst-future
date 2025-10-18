module github.com/xiajason/zervi-basic/basic/backend/cmd/unified-auth

go 1.25.0

replace github.com/jobfirst/jobfirst-core => ../../pkg/jobfirst-core

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/jobfirst/jobfirst-core v0.0.0-00010101000000-000000000000
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	gorm.io/gorm v1.25.5 // indirect
)
