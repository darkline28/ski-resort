package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/darkline28/ski-resort/backend/internal/auth/api"
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
	return api.RegisterUser201JSONResponse{Id: user.ID, Email: user.Email, Name: user.Name, Role: api.RegisterUser201JSONResponseBodyRole(user.Role),
		AccessToken: token, TokenType: api.RegisterUser201JSONResponseBodyTokenType("Bearer")}, nil

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

	return api.LoginUser200JSONResponse{AccessToken: token, TokenType: api.LoginUser200JSONResponseBodyTokenType("Bearer")}, nil

}
