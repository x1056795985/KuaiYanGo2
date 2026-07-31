package captcha

import "github.com/mojocn/base64Captcha"

// GenerateDigit creates a four-digit image challenge valid for five minutes.
func GenerateDigit() (id, dataURL string, err error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	return base64Captcha.NewCaptcha(driver, VerificationCodes).Generate()
}
