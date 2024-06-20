package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/gofiber/fiber/v2"
	"github.com/leonfiasco/boomerang-go/database"
	"github.com/leonfiasco/boomerang-go/models/user"
	"github.com/leonfiasco/boomerang-go/models/userToken"
	"github.com/leonfiasco/boomerang-go/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var verifier = emailverifier.NewVerifier()

const development = "http://localhost:2402"
const verificationEmailTemplate = `
	<p>Verify your email address to complete the signup and login into your account.</p>
	<p>This link <b>expires in 1 hour</b>.</p>
	<p>Press <a href="%v">here</a> to proceed.</p>
`

func generateToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

func Register(c *fiber.Ctx) error {
	var requestBody user.UserRequestBody

	userCollection := database.Mg.DB.Collection("users")
	tokenCollection := database.Mg.DB.Collection("tokens")

	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if requestBody.FirstName == "" || requestBody.LastName == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Firstname and Lastname are required"})
	}

	ret, err := verifier.Verify(requestBody.Email)
	if err != nil || !ret.Syntax.Valid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid email address"})
	}

	if len(requestBody.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password must be at least 6 characters long"})
	}

	var existingUser user.User
	err = userCollection.FindOne(context.Background(), bson.M{"email": requestBody.Email}).Decode(&existingUser)
	if err == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "User with this email already exists"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error hashing password"})
	}

	newUser := user.User{
		FirstName: requestBody.FirstName,
		LastName:  requestBody.LastName,
		Email:     requestBody.Email,
		Password:  string(hashedPassword),
	}

	insertResult, err := userCollection.InsertOne(context.Background(), newUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error saving user"})
	}

	userID := insertResult.InsertedID.(primitive.ObjectID).Hex()

	tokenStr, err := generateToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
	}

	newToken := userToken.Token{
		UserID:    userID,
		Token:     tokenStr,
		CreatedAt: time.Now(),
	}

	url := fmt.Sprintf("%v/user/%v/verify/%v", development, userID, newToken.Token)
	html := fmt.Sprintf(verificationEmailTemplate, url)

	utils.SendEmail("Verify Email", html, []string{newUser.Email})

	_, err = tokenCollection.InsertOne(context.Background(), newToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error saving token"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Verification email has been sent",
		"success":    true,
		"statusCode": fiber.StatusCreated,
		"user":       newUser,
	})
}

func Login(c *fiber.Ctx) error {
	var requestBody user.UserRequestBody

	userCollection := database.Mg.DB.Collection("users")
	tokenCollection := database.Mg.DB.Collection("tokens")

	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	var existingUser user.User
	err := userCollection.FindOne(context.Background(), bson.M{"email": requestBody.Email}).Decode(&existingUser)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(requestBody.Password)) != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	tokenStr, err := generateToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
	}

	newToken := userToken.Token{
		UserID: existingUser.Email,
		Token:  tokenStr,
	}

	_, err = tokenCollection.InsertOne(context.Background(), newToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error saving token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "statusCode": fiber.StatusOK, "token": tokenStr})
}

func VerifyEmail(c *fiber.Ctx) error {
	userCollection := database.Mg.DB.Collection("users")
	tokenCollection := database.Mg.DB.Collection("tokens")

	userID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}
	tokenStr := c.Params("token")

	var existingUser user.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&existingUser)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid user link"})
	}

	var existingToken userToken.Token
	err = tokenCollection.FindOne(context.Background(), bson.M{"userId": userID.Hex(), "token": tokenStr}).Decode(&existingToken)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid token link"})
	}

	currentTime := time.Now()
	tokenExpirationTime := existingToken.CreatedAt.Add(1 * time.Hour)

	if currentTime.After(tokenExpirationTime) {
		_, err := tokenCollection.DeleteOne(context.Background(), bson.M{"_id": existingToken.ID})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error deleting expired token"})
		}
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "Token has expired"})
	}

	_, err = userCollection.UpdateOne(context.Background(), bson.M{"_id": userID}, bson.M{"$set": bson.M{"verified": true}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error updating user verification status"})
	}

	_, err = tokenCollection.DeleteOne(context.Background(), bson.M{"_id": existingToken.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error deleting token"})
	}

	return c.Render("verified", fiber.Map{})
}

func ResendVerification(c *fiber.Ctx) error {
	var requestBody struct {
		Email  string `json:"email" validate:"required,email"`
		UserID string `json:"userId" validate:"required, userId"`
	}

	userCollection := database.Mg.DB.Collection("users")
	tokenCollection := database.Mg.DB.Collection("tokens")

	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	userID, err := primitive.ObjectIDFromHex(requestBody.UserID)

	fmt.Println(userID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	if requestBody.Email == "" || requestBody.UserID == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Empty user details are not allowed"})
	}

	var existingUser user.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&existingUser)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid user link"})
	}

	var existingToken userToken.Token
	err = tokenCollection.FindOne(context.Background(), bson.M{"userId": userID.Hex()}).Decode(&existingToken)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid token link"})
	}

	newTokenValue, err := generateToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
	}

	_, err = tokenCollection.UpdateOne(context.Background(), bson.M{"_id": existingToken.ID}, bson.M{"$set": bson.M{"token": newTokenValue, "createdAt": time.Now()}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error updating token"})
	}

	url := fmt.Sprintf("%v/user/%v/verify/%v", development, requestBody.UserID, newTokenValue)
	html := fmt.Sprintf(verificationEmailTemplate, url)

	utils.SendEmail("Verify Email", html, []string{existingUser.Email})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Verification email has been resent",
	})
}
