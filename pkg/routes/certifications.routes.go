package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

func CertificationRoutes(app *fiber.App, certService *services.CertificationService, authMiddleware *middleware.AuthMiddleware) {
	r := app.Group("/certifications").Use(authMiddleware.Auth)

	r.Get("/categories", func(c *fiber.Ctx) error {
		return c.JSON(config.AllCertificationCategories)
	})

	r.Get("/types", func(c *fiber.Ctx) error {
		category := c.Query("category")

		types := certService.GetTypes(category)

		return c.JSON(types)
	})

	r.Get("/list", func(c *fiber.Ctx) error {
		userId := utils.GetUser(c).ID
		certs, err := certService.ListUserCertifications(userId)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error listing certifications")
		}
		return c.Render("certifications/list", fiber.Map{
			"Certs": certs,
		})
	})

	r.Get("/create", func(c *fiber.Ctx) error {
		return c.Render("certifications/create", fiber.Map{
			"Cert": types.Certification{
				StartDate: types.Date(time.Now()),
				EndDate:   types.Date(time.Now().AddDate(2, 0, 0)),
			},
			"CertificationTypes": config.AllCertificationTypes,
		})
	})

	r.Post("/create", func(c *fiber.Ctx) error {
		userId := utils.GetUser(c).ID

		cert := new(types.Certification)

		if err := utils.ParseBodyAndValidate(c, cert); err != nil {
			return c.Render("certifications/create", fiber.Map{
				"Cert":  cert,
				"error": err.Message,
			})
		}
		cert.UserID = &userId
		cert.ID = uuid.New()

		if _, err := certService.CreateCertification(cert); err != nil {
			return c.Render("certifications/create", fiber.Map{
				"Cert":  cert,
				"error": err.Error(),
			})
		}
		return c.Redirect("/certifications/" + cert.ID.String())
	})

	r.Post("/:certId/edit", func(c *fiber.Ctx) error {
		userId := utils.GetUser(c).ID

		certId, err := uuid.Parse(c.Params("certId"))
		if err != nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid Certification ID")
		}

		cert := new(types.Certification)
		if err := utils.ParseBodyAndValidate(c, cert); err != nil {
			return err
		}
		cert.ID = certId

		if _, err := certService.UpdateCertification(userId, cert); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error updating certification")
		}
		return c.Redirect("/certifications/" + cert.ID.String())
	})

	r.Post("/:certId/delete", func(c *fiber.Ctx) error {
		userId := utils.GetUser(c).ID
		certId, err := uuid.Parse(c.Params("certId"))
		if err != nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid Certification ID")
		}
		err = certService.DeleteCertification(certId, (userId))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error deleting certification")
		}
		return c.Redirect("/certifications/list")
	})

	r.Get("/:certId", func(c *fiber.Ctx) error {
		userId := utils.GetUser(c).ID
		certId, err := uuid.Parse(c.Params("certId"))
		if err != nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid Certification ID")
		}
		cert, err := certService.GetCertification(userId, certId)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error getting certification")
		}
		return c.Render("certifications/show", fiber.Map{
			"Cert": cert,
		})
	})

	r.Get("/:certId/edit", func(c *fiber.Ctx) error {
		userId := utils.GetUser(c).ID
		certId, err := uuid.Parse(c.Params("certId"))
		if err != nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid Certification ID")
		}
		cert, err := certService.GetCertification(userId, certId)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error getting certification")
		}
		return c.Render("certifications/edit", fiber.Map{
			"Cert": cert,
		})
	})

}
