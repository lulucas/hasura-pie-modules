package account

type Captcha interface {
	ValidateImageCaptcha(id, captcha string) error
	ValidateSmsCaptcha(mobile, captcha string) error
}
