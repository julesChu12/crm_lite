package main

import (
	"crm_lite/cmd"
	"os"
)

// @title           CRM Lite API
// @version         1.0
// @description     This is a lightweight CRM system.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
