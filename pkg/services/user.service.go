package services

import (
	"fmt"
	"html/template"
	"log/slog"
	"strconv"
	"strings"
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

// categoryColors maps category IDs to hex colors (matching the design palette).
var categoryColors = map[string]string{
	"climbing": "#1899b0",
	"sailing":  "#5cba7d",
	"camping":  "#f3c01c",
	"foraging": "#e8590c",
	"other":    "#a9d3e8",
}

type CategoryLegendItem struct {
	Name  string
	Color template.CSS
}

// heatmapLegend returns per-category legend items plus a "Flere kategorier" entry.
func heatmapLegend() []CategoryLegendItem {
	items := make([]CategoryLegendItem, 0, len(types.AllActivityCategories)+1)
	for _, cat := range types.AllActivityCategories {
		hex := categoryColors[cat.ID]
		if hex == "" {
			hex = "#a9d3e8"
		}
		items = append(items, CategoryLegendItem{Name: cat.Name, Color: template.CSS(hex)})
	}
	// Gradient swatch for months with multiple categories
	items = append(items, CategoryLegendItem{
		Name:  "Flere kategorier",
		Color: template.CSS("linear-gradient(135deg,#1899b0 0 50%,#5cba7d 50% 100%)"),
	})
	return items
}

type ActivityHeatmapCell struct {
	Count         int
	Color         template.CSS // solid color or CSS gradient; empty = no activity
	IsCurrentPeriod bool
	Tooltip       string
}

type ActivityHeatmapRow struct {
	Year   int
	Months [12]ActivityHeatmapCell
}

// categoryGradient builds a CSS background value for the active categories in
// AllActivityCategories order: solid for one category, linear for two, conic for three+.
func categoryGradient(catCounts map[string]int) template.CSS {
	colors := make([]string, 0, len(catCounts))
	for _, cat := range types.AllActivityCategories {
		if catCounts[cat.ID] > 0 {
			c := categoryColors[cat.ID]
			if c == "" {
				c = "#a9d3e8"
			}
			colors = append(colors, c)
		}
	}
	switch len(colors) {
	case 0:
		return ""
	case 1:
		return template.CSS(colors[0])
	case 2:
		return template.CSS(fmt.Sprintf("linear-gradient(135deg,%s 0 50%%,%s 50%% 100%%)", colors[0], colors[1]))
	default:
		pct := 100.0 / float64(len(colors))
		parts := make([]string, len(colors))
		for i, c := range colors {
			start := float64(i) * pct
			end := float64(i+1) * pct
			if i == len(colors)-1 {
				end = 100
			}
			parts[i] = fmt.Sprintf("%s %.4g%% %.4g%%", c, start, end)
		}
		return template.CSS("conic-gradient(" + strings.Join(parts, ",") + ")")
	}
}

// buildCell creates an ActivityHeatmapCell from per-category counts and a date label.
func buildCell(catCounts map[string]int, total int, dateLabel string, isCurrent bool) ActivityHeatmapCell {
	if total == 0 {
		return ActivityHeatmapCell{Tooltip: dateLabel + ": ingen aktiviteter", IsCurrentPeriod: isCurrent}
	}

	tooltipParts := make([]string, 0, len(catCounts))
	for _, cat := range types.AllActivityCategories {
		if c := catCounts[cat.ID]; c > 0 {
			tooltipParts = append(tooltipParts, cat.Name+" ("+strconv.Itoa(c)+")")
		}
	}

	return ActivityHeatmapCell{
		Count:         total,
		Color:         categoryGradient(catCounts),
		IsCurrentPeriod: isCurrent,
		Tooltip:       dateLabel + ": " + strings.Join(tooltipParts, ", "),
	}
}

// GetUserActivityHeatmap returns one row per year with monthly and weekly cells.
func (s *UserService) GetUserActivityHeatmap(user *types.User) []ActivityHeatmapRow {
	if len(user.Activities) == 0 {
		return nil
	}

	now := time.Now()
	type cellKey = [2]int
	monthCats := make(map[cellKey]map[string]int)
	monthTotals := make(map[cellKey]int)
	minYear := user.Activities[0].Date.Time().Year()
	maxYear := now.Year()

	for _, act := range user.Activities {
		t := act.Date.Time()
		if t.Year() < minYear {
			minYear = t.Year()
		}
		mk := cellKey{t.Year(), int(t.Month())}
		if monthCats[mk] == nil {
			monthCats[mk] = make(map[string]int)
		}
		monthCats[mk][act.Category.ID]++
		monthTotals[mk]++
	}

	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "Maj", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dec"}
	rows := make([]ActivityHeatmapRow, 0, maxYear-minYear+1)
	for y := minYear; y <= maxYear; y++ {
		row := ActivityHeatmapRow{Year: y}
		for m := 1; m <= 12; m++ {
			mk := cellKey{y, m}
			isCurrent := y == now.Year() && m == int(now.Month())
			row.Months[m-1] = buildCell(monthCats[mk], monthTotals[mk], monthNames[m-1]+" "+strconv.Itoa(y), isCurrent)
		}
		rows = append(rows, row)
	}
	return rows
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

	activityHeatmap := s.GetUserActivityHeatmap(user)

	return ctx.Render("users/user", fiber.Map{
		"User":            *user,
		"ActivityHeatmap": activityHeatmap,
		"HeatmapLegend":   heatmapLegend(),
		"Categories":      types.AllActivityCategories,
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
