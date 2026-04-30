package api

const (
	// auth
	ErrUserNotFound         = "user does not exist"
	ErrUserAlreadyExists    = "username or email already taken"
	ErrEmailAlreadyVerified = "email already verified, proceed to login"
	ErrInvalidCredentials   = "invalid email or password"
	ErrUnauthorized         = "unauthorized"

	// otp
	ErrOTPNotFound   = "valid OTP not found, request a new one"
	ErrOTPInvalid    = "invalid OTP code"
	ErrOTPCooldown   = "wait before requesting another OTP"
	ErrGeneratingOTP = "error generating OTP"
	ErrSendingOTP    = "error sending OTP"

	// session
	ErrSessionNotFound = "session not found"
	ErrSessionRevoked  = "session has been revoked"
	ErrSessionExpired  = "session has expired"

	// token
	ErrInvalidToken = "invalid token"

	// game
	ErrGameNotFound      = "game not found"
	ErrGameNotActive     = "game is not active"
	ErrNotAPlayer        = "you are not a player in this game"
	ErrCannotJoinOwnGame = "cannot join your own game"
	ErrGameNotJoinable   = "game is not available to join"

	// voice
	ErrVoiceSessionNotFound      = "voice session not found"
	ErrVoiceSessionAlreadyActive = "a voice call is already in progress"

	// general
	ErrInternalServer = "an error occurred"
	ErrInvalidInput   = "invalid input"
)
