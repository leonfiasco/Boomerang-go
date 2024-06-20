package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leonfiasco/boomerang-go/controller/user"
)

func SetupRoutes(userGroup fiber.Router) {

	userGroup.Post("/register", user.Register)
	userGroup.Post("/login", user.Login)
	userGroup.Get("/:id/verify/:token", user.VerifyEmail)
	userGroup.Post("/resendVerification", user.ResendVerification)

}
