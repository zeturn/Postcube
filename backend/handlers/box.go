package handlers

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/WorkPlace/Postcube/backend/database"
	"github.com/WorkPlace/Postcube/backend/models"
	"github.com/gofiber/fiber/v2"
)

var colorPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

type questionResponse struct {
	ID              uint       `json:"id"`
	AnonymousName   string     `json:"anonymous_name"`
	Content         string     `json:"content"`
	Answer          string     `json:"answer"`
	Status          string     `json:"status"`
	BackgroundColor string     `json:"background_color"`
	CreatedAt       time.Time  `json:"created_at"`
	AnsweredAt      *time.Time `json:"answered_at"`
}

func questionToResponse(q models.Question) questionResponse {
	return questionResponse{
		ID:              q.ID,
		AnonymousName:   q.AnonymousName,
		Content:         q.Content,
		Answer:          q.Answer,
		Status:          q.Status,
		BackgroundColor: q.BackgroundColor,
		CreatedAt:       q.CreatedAt,
		AnsweredAt:      q.AnsweredAt,
	}
}

func GetPublicBox(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid slug"})
	}

	var user models.User
	if err := database.DB.Where("slug = ?", slug).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "box not found"})
	}

	var questions []models.Question
	if err := database.DB.Where("owner_id = ?", user.ID).Order("created_at desc").Find(&questions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load box"})
	}

	answered := make([]questionResponse, 0)
	unanswered := make([]questionResponse, 0)
	for _, q := range questions {
		resp := questionToResponse(q)
		if q.Status == models.QuestionStatusAnswered {
			answered = append(answered, resp)
		} else {
			unanswered = append(unanswered, resp)
		}
	}

	return c.JSON(fiber.Map{
		"owner": fiber.Map{
			"name":      user.Name,
			"slug":      user.Slug,
			"box_title": user.BoxTitle,
		},
		"answered":   answered,
		"unanswered": unanswered,
	})
}

func SubmitAnonymousQuestion(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid slug"})
	}

	var user models.User
	if err := database.DB.Where("slug = ?", slug).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "box not found"})
	}

	var body struct {
		Content       string `json:"content"`
		AnonymousName string `json:"anonymous_name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	content := strings.TrimSpace(body.Content)
	if content == "" || len(content) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "content must be 1-500 chars"})
	}

	name := strings.TrimSpace(body.AnonymousName)
	if name == "" {
		name = "Anonymous"
	}
	if len(name) > 24 {
		name = name[:24]
	}

	question := models.Question{
		OwnerID:         user.ID,
		AnonymousName:   name,
		Content:         content,
		Status:          models.QuestionStatusPending,
		BackgroundColor: "#fff4d6",
	}
	if err := database.DB.Create(&question).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create question"})
	}

	return c.Status(fiber.StatusCreated).JSON(questionToResponse(question))
}

func GetMyBox(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	var total int64
	var answered int64
	database.DB.Model(&models.Question{}).Where("owner_id = ?", userID).Count(&total)
	database.DB.Model(&models.Question{}).Where("owner_id = ? AND status = ?", userID, models.QuestionStatusAnswered).Count(&answered)

	return c.JSON(fiber.Map{
		"user": user,
		"stats": fiber.Map{
			"total":      total,
			"answered":   answered,
			"unanswered": total - answered,
		},
	})
}

func UpdateMyBox(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var body struct {
		BoxTitle string `json:"box_title"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	title := strings.TrimSpace(body.BoxTitle)
	if title == "" || len(title) > 120 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "box_title must be 1-120 chars"})
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	user.BoxTitle = title
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update box"})
	}

	return c.JSON(user)
}

func GetInbox(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var questions []models.Question
	if err := database.DB.Where("owner_id = ?", userID).Order("created_at desc").Find(&questions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load inbox"})
	}

	result := make([]questionResponse, 0, len(questions))
	for _, q := range questions {
		result = append(result, questionToResponse(q))
	}
	return c.JSON(result)
}

func UpdateInboxQuestion(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid question id"})
	}

	var question models.Question
	if err := database.DB.Where("id = ? AND owner_id = ?", id, userID).First(&question).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "question not found"})
	}

	var body struct {
		Answer          *string `json:"answer"`
		BackgroundColor *string `json:"background_color"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if body.BackgroundColor != nil {
		color := strings.TrimSpace(*body.BackgroundColor)
		if !colorPattern.MatchString(color) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "background_color must be hex color"})
		}
		question.BackgroundColor = strings.ToLower(color)
	}

	if body.Answer != nil {
		trimmed := strings.TrimSpace(*body.Answer)
		question.Answer = trimmed
		if trimmed == "" {
			question.Status = models.QuestionStatusPending
			question.AnsweredAt = nil
		} else {
			now := time.Now()
			question.Status = models.QuestionStatusAnswered
			question.AnsweredAt = &now
		}
	}

	if err := database.DB.Save(&question).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update question"})
	}

	return c.JSON(questionToResponse(question))
}

func DeleteInboxQuestion(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid question id"})
	}

	result := database.DB.Where("id = ? AND owner_id = ?", id, userID).Delete(&models.Question{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete question"})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "question not found"})
	}

	return c.JSON(fiber.Map{"message": "deleted"})
}
