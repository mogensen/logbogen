package services

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"
)

// UserService handles authentication related operations
type UserService struct {
	userDal dal.UserDal
}

// NewUserService creates a new instance of UserService
func NewUserService(userDal dal.UserDal) *UserService {
	return &UserService{
		userDal: userDal,
	}
}

// GetUsers returns all users except the current user
func (s *UserService) GetUsers(currentUserID uint64) (*GetUsersResponse, error) {
	users, err := s.userDal.FindUsers()
	if err != nil {
		return nil, err
	}

	res := make([]*types.UserForLogin, 0, len(users))
	for _, v := range users {
		user := v
		if uint64(v.ID) == currentUserID {
			continue // Skip the current user
		}

		res = append(res, &types.UserForLogin{
			ID:    uint64(user.ID),
			Name:  user.Name,
			Email: user.Email,
		})
	}

	return &GetUsersResponse{
		Users: res,
	}, nil
}

var activityCategoryPalette = [5]string{"#219ebc", "#023047", "#ffb703", "#fb8500", "#8ecae6"}

type ActivityMonthCategoryBar struct {
	Year       int
	Month      int
	Label      string                    // MM/YYYY
	Categories []ActivityCategorySegment // ordered by main categories
	Total      int
}

type ActivityCategorySegment struct {
	CategoryID   string
	CategoryName string
	Color        string
	Count        int
	HeightPct    float64 // 0-100, proportional height for the bar
	Tooltip      string  // "MM/YYYY: CategoryName (Count)"
}

// GetUserActivityBars returns a slice of ActivityMonthCategoryBar for the user's activities
func (s *UserService) GetUserActivityBars(user *types.User) []ActivityMonthCategoryBar {
	categories := types.AllActivityCategories
	categoryIDs := make([]string, len(categories))
	for i, cat := range categories {
		categoryIDs[i] = cat.ID
	}

	if len(user.Activities) == 0 {
		return nil
	}
	minDate := user.Activities[0].Date.Time()
	maxDate := time.Now()
	for _, act := range user.Activities {
		if act.Date.Time().Before(minDate) {
			minDate = act.Date.Time()
		}
	}

	bars := []ActivityMonthCategoryBar{}
	for d := minDate; !d.After(maxDate); d = d.AddDate(0, 1, 0) {
		counts := make(map[string]int)
		total := 0
		for _, catID := range categoryIDs {
			counts[catID] = 0
		}
		for _, act := range user.Activities {
			if act.Date.Time().Year() == d.Year() && act.Date.Time().Month() == d.Month() {
				counts[act.Category.ID]++
				total++
			}
		}
		label := d.Format("01/2006")
		segments := make([]ActivityCategorySegment, len(categoryIDs))
		for i, cat := range categories {
			count := counts[cat.ID]
			pct := 0.0
			if total > 0 {
				pct = float64(count) * 20
			}
			color := ""
			if i < len(activityCategoryPalette) {
				color = activityCategoryPalette[i]
			}
			tooltip := label + ": " + cat.Name + " (" + strconv.Itoa(count) + ")"
			segments[i] = ActivityCategorySegment{
				CategoryID:   cat.ID,
				CategoryName: cat.Name,
				Color:        color,
				Count:        count,
				HeightPct:    pct,
				Tooltip:      tooltip,
			}
		}
		bars = append(bars, ActivityMonthCategoryBar{
			Year:       d.Year(),
			Month:      int(d.Month()),
			Label:      label,
			Categories: segments,
			Total:      total,
		})
	}
	return bars
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(userID uint64) (*types.User, error) {
	user, err := s.userDal.FindUserById(userID)
	if err != nil {
		return nil, err
	}

	return types.UserFromDal(user), nil
}

// GetUsersHandler handles the get users HTTP request
func (s *UserService) GetUsersHandler(ctx *fiber.Ctx) error {
	currentUser := utils.GetUser(ctx)
	resp, err := s.GetUsers(currentUser.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return ctx.JSON(resp.Users)
}

// GetUserHandler handles the get user HTTP request
func (s *UserService) GetUserHandler(ctx *fiber.Ctx) error {
	userIdParam := ctx.Params("UserID")
	if userIdParam == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid user")
	}

	userId, err := strconv.ParseUint(userIdParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid user")
	}

	user, err := s.GetUserByID(userId)
	if err != nil {
		slog.Error("Failed to get user by ID", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	activityBars := s.GetUserActivityBars(user)

	return ctx.Render("users/user", fiber.Map{
		"User":         *user,
		"ActivityBars": activityBars,
		"Categories":   types.AllActivityCategories,
	})
}

// UpdateUser updates a user's properties
func (s *UserService) UpdateUser(user *dal.User) error {
	return s.userDal.UpdateUser(user)
}

// UpdateThemeRequest represents the request body for theme updates
type UpdateThemeRequest struct {
	Theme string `json:"theme"`
}

// UpdateThemeHandler handles POST requests to update the user's theme preference
func (s *UserService) UpdateThemeHandler(ctx *fiber.Ctx) error {
	currentUser := utils.GetUser(ctx)
	if currentUser == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Not authenticated")
	}

	var req UpdateThemeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate theme value
	if req.Theme != "light" && req.Theme != "dark" && req.Theme != "auto" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid theme. Must be 'light', 'dark', or 'auto'")
	}

	// Get the DAL user and update
	user, err := s.userDal.FindUserById(uint64(currentUser.ID))
	if err != nil {
		slog.Error("Failed to find user for theme update", "error", err, "userID", currentUser.ID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update theme")
	}

	user.ThemePref = req.Theme
	if err := s.userDal.UpdateUser(user); err != nil {
		slog.Error("Failed to update user theme", "error", err, "userID", currentUser.ID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update theme")
	}

	// Update the current user in context to reflect the change
	currentUser.ThemePref = req.Theme

	return ctx.JSON(fiber.Map{
		"theme": req.Theme,
		"message": "Theme updated successfully",
	})
}
