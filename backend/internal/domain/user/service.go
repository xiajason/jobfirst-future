package user

import "context"

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	GetProfile(ctx context.Context, userID uint) (*User, error)
	UpdateProfile(ctx context.Context, userID uint, req UpdateProfileRequest) (*User, error)
	ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest) error
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	VerifyEmail(ctx context.Context, req VerifyEmailRequest) error
	VerifyPhone(ctx context.Context, req VerifyPhoneRequest) error
	List(ctx context.Context, req ListRequest) (*ListResponse, error)
	UpdateStatus(ctx context.Context, userID uint, status Status) error
}
