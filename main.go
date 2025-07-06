package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/gomail.v2"
)

type EmailRequest struct {
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Body    string `json:"body" validate:"required"`
}

type EmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.POST("/send-mail", sendMailHandler)
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Mail Service API",
			"version": "1.0.0",
		})
	})

	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

func sendMailHandler(c echo.Context) error {
	var req EmailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, EmailResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	if req.To == "" || req.Subject == "" || req.Body == "" {
		return c.JSON(http.StatusBadRequest, EmailResponse{
			Success: false,
			Message: "Missing required fields: to, subject, body",
		})
	}

	err := sendEmail(req.To, req.Subject, req.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, EmailResponse{
			Success: false,
			Message: "Failed to send email: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, EmailResponse{
		Success: true,
		Message: "Email sent successfully",
	})
}

func sendEmail(to string, subject string, body string) error {
	// Get SMTP configuration from environment
	host := os.Getenv("MAIL_HOST")
	portStr := os.Getenv("MAIL_PORT")
	username := os.Getenv("MAIL_USERNAME")
	password := os.Getenv("MAIL_PASSWORD")
	fromAddress := os.Getenv("MAIL_FROM_ADDRESS")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	// Create message
	m := gomail.NewMessage()
	m.SetHeader("From", fromAddress)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Configure SMTP dialer
	d := gomail.NewDialer(host, port, username, password)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
