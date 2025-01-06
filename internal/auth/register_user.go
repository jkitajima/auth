package auth

import (
	"context"

	"auth/internal/user"

	"github.com/alexedwards/argon2id"
)

type RegisterRequest struct {
	Email    string
	Password string
}

type RegisterResponse struct {
	User *user.User
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	// TODO: Password strength policy validation

	// Hash password
	hashedPasswd, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams)
	if err != nil {
		return RegisterResponse{nil}, err
	}

	// Generate OTP and send email for validation
	// otp, err := generateOTP()
	// if err != nil {
	// 	return RegisterResponse{nil}, err
	// }
	// otpExpiration := int(time.Now().Add(5 * time.Minute).Unix())

	user := &user.User{
		Email:    req.Email,
		Password: hashedPasswd,
		// VerificationCode:           &otp,
		// VerificationCodeExpiration: &otpExpiration,
	}
	err = s.UserRepo.Insert(ctx, user)
	if err != nil {
		return RegisterResponse{nil}, err
	}
	return RegisterResponse{user}, err
}

// func generateOTP() (otp string, err error) {
// 	length := 6
// 	maxInt := big.NewInt(9)

// 	for i := 0; i < length; i++ {
// 		number, err := rand.Int(rand.Reader, maxInt)
// 		if err != nil {
// 			return "", err
// 		}
// 		otp += number.String()
// 	}

// 	return otp, nil
// }
