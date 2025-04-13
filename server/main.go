package main

import (
	// "net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Start(":5001")
}
