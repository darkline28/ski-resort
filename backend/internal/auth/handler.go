package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/darkline28/ski-resort/backend/internal/auth/api"
	"github.com/darkline28/ski-resort/backend/internal/middleware"
	"github.com/oapi-codegen/runtime/types"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterUser implements [api.StrictServerInterface].
func (h *Handler) RegisterUser(ctx context.Context, request api.RegisterUserRequestObject) (api.RegisterUserResponseObject, error) {
	if request.Body == nil {
		return api.RegisterUser400JSONResponse{Error: "invalid_json", Message: "Неверный запрос"}, nil
	}
	input := RegisterInput{
		Name:     request.Body.Name,
		Email:    string(request.Body.Email),
		Phone:    request.Body.Phone,
		Password: request.Body.Password,
	}

	user, token, err := h.service.Register(ctx, input)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			return api.RegisterUser400JSONResponse{Error: "invalid_input", Message: "Некорректные данные регистрации"}, nil
		}
		if errors.Is(err, ErrEmailAlreadyExists) {
			return api.RegisterUser409JSONResponse{Error: "email_already_exists", Message: "Данный email уже используется"}, nil
		}
		slog.ErrorContext(ctx, "registration failed", "error", err)
		return api.RegisterUser500JSONResponse{Error: "internal_error", Message: "Внутренняя ошибка сервера"}, nil
	}
	slog.InfoContext(ctx,
		"user registered",
		"user_id", user.ID,
		"email", user.Email,
	)
	response := api.RegisterUser201JSONResponse{
		AccessToken: token,
		TokenType:   api.RegisterUser201JSONResponseBodyTokenType("Bearer"),
	}
	response.User.Id = user.ID
	response.User.Email = user.Email
	response.User.Name = user.Name
	response.User.Phone = user.Phone
	response.User.Role = api.RegisterUser201JSONResponseBodyUserRole(user.Role)

	return response, nil
}

// LoginUser implements [api.StrictServerInterface].
func (h *Handler) LoginUser(ctx context.Context, request api.LoginUserRequestObject) (api.LoginUserResponseObject, error) {
	if request.Body == nil {
		return api.LoginUser400JSONResponse{Error: "invalid_json", Message: "Неверный запрос"}, nil
	}

	email := request.Body.Email
	password := request.Body.Password
	user, token, err := h.service.Login(ctx, string(email), password)

	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			return api.LoginUser400JSONResponse{Error: "invalid_input", Message: "Некорректные данные авторизации"}, nil
		}
		if errors.Is(err, ErrInvalidEmailOrPassword) {
			return api.LoginUser401JSONResponse{Error: "invalid_password_or_email", Message: "Неверный email или пароль"}, nil
		}
		slog.ErrorContext(ctx, "login failed", "error", err)
		return api.LoginUser500JSONResponse{Error: "internal_error", Message: "Внутренняя ошибка сервера"}, nil
	}

	slog.InfoContext(ctx, "user logged in", "user_id", user.ID)

	response := api.LoginUser200JSONResponse{
		AccessToken: token,
		TokenType:   api.LoginUser200JSONResponseBodyTokenType("Bearer"),
	}

	response.User.Id = user.ID
	response.User.Name = user.Name
	response.User.Email = user.Email
	response.User.Role = api.LoginUser200JSONResponseBodyUserRole(user.Role)

	if user.Phone != "" {
		response.User.Phone = &user.Phone
	}

	return response, nil
}

// GetUser implements [api.StrictServerInterface].
func (h *Handler) GetUser(ctx context.Context, request api.GetUserRequestObject) (api.GetUserResponseObject, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return api.GetUser401JSONResponse{Error: "unauthorized", Message: "Пользователь не авторизован"}, nil
	}

	user, err := h.service.GetCurrentUser(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return api.GetUser401JSONResponse{Error: "unauthorized", Message: "Пользователь не авторизован"}, nil
		}
		slog.ErrorContext(ctx, "getUser failed", "error", err)
		return api.GetUser500JSONResponse{Error: "internal_error", Message: "Внутренняя ошибка сервера"}, nil
	}
	response := api.GetUser200JSONResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: types.Email(user.Email),
		Phone: user.Phone,
		Role:  api.GetUser200JSONResponseBodyRole(user.Role),
	}

	return response, nil
}
