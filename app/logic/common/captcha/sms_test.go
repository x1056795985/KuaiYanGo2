package captcha

import "testing"

func TestValidateSMSInput(t *testing.T) {
	tests := []struct {
		name      string
		variables []string
		phone     string
		wantError bool
	}{
		{name: "valid", variables: []string{"123456"}, phone: "13800138000"},
		{name: "missing variables", phone: "13800138000", wantError: true},
		{name: "missing phone", variables: []string{"123456"}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateSMSInput(test.variables, test.phone)
			if (err != nil) != test.wantError {
				t.Fatalf("validateSMSInput() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestAliyunURLEncode(t *testing.T) {
	if got, want := aliyunURLEncode("a b+c*~"), "a%20b%2Bc%2A~"; got != want {
		t.Fatalf("aliyunURLEncode() = %q, want %q", got, want)
	}
}
