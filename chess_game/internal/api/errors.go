package api

// auth errors
const (
	ErrUserNotFound         = "user does not exist"
	ErrUserAlreadyExists    = "username or email already taken"
	ErrEmailAlreadyVerified = "email already verified, proceed to login"
	ErrInvalidCredentials   = "invalid email or password"
	ErrUnauthorized         = "unauthorized"

	// otp errors
	ErrOTPNotFound   = "valid OTP not found, request a new one"
	ErrOTPInvalid    = "invalid OTP code"
	ErrOTPCooldown   = "wait before requesting another OTP"
	ErrGeneratingOTP = "error generating OTP"
	ErrSendingOTP    = "error sending OTP"

	// session errors
	ErrSessionNotFound = "session not found"
	ErrSessionRevoked  = "session has been revoked"
	ErrSessionExpired  = "session has expired"

	//token errors
	ErrInvalidToken = "invalid token"

	// general
	ErrInternalServer = "an error occurred"
	ErrInvalidInput   = "invalid input"
)
