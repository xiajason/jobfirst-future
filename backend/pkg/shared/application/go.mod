module resume-centre/shared/application

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/viper v1.17.0
	resume-centre/shared/kernel v0.0.0
	resume-centre/shared/infrastructure v0.0.0
)

replace (
	resume-centre/shared/kernel => ../kernel
	resume-centre/shared/infrastructure => ../infrastructure
)
