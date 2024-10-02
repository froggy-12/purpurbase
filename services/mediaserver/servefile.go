package mediaserver

import (
	"os"
	"path/filepath"

	"github.com/froggy-12/purpurbase/types"
	"github.com/gofiber/fiber/v2"
)

func ServeFiles(c *fiber.Ctx) error {
	folder := c.Query("folder")
	fileName := c.Query("file_name")
	if folder == "" || fileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "pass right queries please"})
	}
	uploadDir := filepath.Join(".", "uploads", folder)

	requestedFile := filepath.Join(uploadDir, fileName)

	_, err := os.Stat(requestedFile)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(types.ErrorResponse{Error: "File not found"})
	}

	return c.SendFile(requestedFile)
}

func DownloadFile(c *fiber.Ctx) error {
	folder := c.Query("folder")
	fileName := c.Query("file_name")
	if folder == "" || fileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "pass right queries please"})
	}
	uploadDir := filepath.Join(".", "uploads", folder)

	requestedFile := filepath.Join(uploadDir, fileName)

	_, err := os.Stat(requestedFile)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(types.ErrorResponse{Error: "File not found"})
	}

	return c.Download(requestedFile, fileName)
}
