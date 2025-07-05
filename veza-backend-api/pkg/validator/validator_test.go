package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStruct pour tester les validations
type TestStruct struct {
	Username    string `json:"username" validate:"required,safe_username"`
	Email       string `json:"email" validate:"required,safe_email"`
	Password    string `json:"password" validate:"required,secure_password"`
	Description string `json:"description" validate:"max_length=500,no_xss"`
	Filename    string `json:"filename" validate:"safe_filename"`
	URL         string `json:"url" validate:"safe_url"`
	Content     string `json:"content" validate:"no_sql_injection,no_special_chars"`
}

// TestStructPassword pour tester uniquement les mots de passe
type TestStructPassword struct {
	Password string `json:"password" validate:"required,secure_password"`
}

// TestStructUsername pour tester uniquement les noms d'utilisateur
type TestStructUsername struct {
	Username string `json:"username" validate:"required,safe_username"`
}

// TestStructEmail pour tester uniquement les emails
type TestStructEmail struct {
	Email string `json:"email" validate:"required,safe_email"`
}

// TestStructContent pour tester uniquement le contenu
type TestStructContent struct {
	Content string `json:"content" validate:"no_sql_injection,no_special_chars"`
}

// TestStructDescription pour tester uniquement les descriptions
type TestStructDescription struct {
	Description string `json:"description" validate:"max_length=500,no_xss"`
}

// TestStructFilename pour tester uniquement les noms de fichiers
type TestStructFilename struct {
	Filename string `json:"filename" validate:"safe_filename"`
}

// TestStructURL pour tester uniquement les URLs
type TestStructURL struct {
	URL string `json:"url" validate:"safe_url"`
}

func TestValidator_ValidateSecurePassword(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "SecurePass123!", false},
		{"Valid complex password", "MyStr0ng!Pass", false},
		{"Too short", "Pass1!", true},
		{"Too long", "ThisPasswordIsWayTooLongForOurSystemAndShouldBeRejectedByTheValidationLogicBecauseItExceedsTheMaximumAllowedLength!", true},
		{"No uppercase", "password123!", true},
		{"No lowercase", "PASSWORD123!", true},
		{"No number", "Password!", true},
		{"No special char", "Password123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStructPassword{
				Password: tt.password,
			}

			err := v.Validate(testStruct)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateSafeUsername(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Valid username", "testuser", false},
		{"Valid with numbers", "test123", false},
		{"Valid with underscore", "test_user", false},
		{"Valid with dash", "test-user", false},
		{"Too short", "ab", true},
		{"Too long", "this_username_is_way_too_long_for_our_system", true},
		{"Invalid chars", "test@user", true},
		{"Invalid chars space", "test user", true},
		{"Consecutive underscores", "test__user", true},
		{"Consecutive dashes", "test--user", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStructUsername{
				Username: tt.username,
			}

			err := v.Validate(testStruct)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateSafeEmail(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "test@example.com", false},
		{"Valid with subdomain", "test@mail.example.com", false},
		{"Valid with numbers", "test123@example.com", false},
		{"Valid with dots", "test.user@example.com", false},
		{"Invalid format no @", "testexample.com", true},
		{"Invalid format no domain", "test@", true},
		{"Invalid format no TLD", "test@example", true},
		{"Invalid chars", "test@exam ple.com", true},
		{"Too long", "verylongemailaddress@verylongdomainname.comverylongemailaddress@verylongdomainname.comverylongemailaddress@verylongdomainname.com", true},
		{"XSS attempt", "test<script>@example.com", true},
		{"SQL injection attempt", "test'OR'1'='1@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStructEmail{
				Email: tt.email,
			}

			err := v.Validate(testStruct)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateNoSQLInjection(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"Valid content", "This is normal content", false},
		{"SQL SELECT", "SELECT * FROM users", true},
		{"SQL INSERT", "INSERT INTO users VALUES", true},
		{"SQL UNION", "UNION SELECT password", true},
		{"SQL comment", "test--comment", true},
		{"SQL block comment", "test/*comment*/", true},
		{"SQL exec", "EXEC sp_help", true},
		{"SQL declare", "DECLARE @var", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStructContent{
				Content: tt.content,
			}

			err := v.Validate(testStruct)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateNoXSS(t *testing.T) {
	v := New()

	tests := []struct {
		name        string
		description string
		wantErr     bool
	}{
		{"Valid description", "This is a normal description", false},
		{"XSS script tag", "test<script>alert('xss')</script>", true},
		{"XSS javascript protocol", "javascript:alert('xss')", true},
		{"XSS onload", "test onload=alert('xss')", true},
		{"XSS onclick", "test onclick=alert('xss')", true},
		{"XSS iframe", "test<iframe src=javascript:alert('xss')>", true},
		{"XSS expression", "test expression(alert('xss'))", true},
		{"XSS eval", "test eval('alert(\"xss\")')", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStructDescription{
				Description: tt.description,
			}

			err := v.Validate(testStruct)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateSafeFilename(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"Valid filename", "document.pdf", false},
		{"Valid with numbers", "file123.txt", false},
		{"Valid with underscore", "my_file.jpg", false},
		{"Valid with dash", "my-file.png", false},
		{"Invalid slash", "path/file.txt", true},
		{"Invalid backslash", "path\\file.txt", true},
		{"Invalid colon", "file:name.txt", true},
		{"Invalid asterisk", "file*.txt", true},
		{"Invalid question mark", "file?.txt", true},
		{"Invalid quotes", "file\".txt", true},
		{"Invalid less than", "file<.txt", true},
		{"Invalid greater than", "file>.txt", true},
		{"Invalid pipe", "file|.txt", true},
		{"System file CON", "CON.txt", true},
		{"System file PRN", "PRN.txt", true},
		{"System file AUX", "AUX.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStructFilename{
				Filename: tt.filename,
			}

			err := v.Validate(testStruct)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateSafeURL(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid HTTP", "http://example.com", false},
		{"Valid HTTPS", "https://example.com", false},
		{"Valid FTP", "ftp://example.com", false},
		{"Valid SFTP", "sftp://example.com", false},
		{"Invalid protocol", "javascript:alert('xss')", true},
		{"Invalid protocol", "data:text/html,<script>alert('xss')</script>", true},
		{"XSS in URL", "http://example.com<script>alert('xss')</script>", true},
		{"SQL injection in URL", "http://example.com'OR'1'='1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStructURL{
				URL: tt.url,
			}

			err := v.Validate(testStruct)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateNoSpecialChars(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"Valid content", "This is normal content", false},
		{"Special chars", "test<script>alert('xss')</script>", true},
		{"Quotes", "test\"quotes\"", true},
		{"Ampersand", "test&content", true},
		{"Semicolon", "test;content", true},
		{"Pipe", "test|content", true},
		{"Backtick", "test`content", true},
		{"Dollar", "test$content", true},
		{"Parentheses", "test(content)", true},
		{"Braces", "test{content}", true},
		{"Brackets", "test[content]", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStructContent{
				Content: tt.content,
			}

			err := v.Validate(testStruct)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_GetValidationErrors(t *testing.T) {
	v := New()

	testStruct := TestStruct{
		Username: "a",                                 // Too short
		Email:    "invalid-email",                     // Invalid format
		Password: "weak",                              // Too weak
		Content:  "test<script>alert('xss')</script>", // XSS attempt
	}

	err := v.Validate(testStruct)
	assert.Error(t, err)

	validationErrors := v.GetValidationErrors(err)
	assert.NotEmpty(t, validationErrors)

	// Vérifier que nous avons des erreurs pour les champs invalides
	errorFields := make(map[string]bool)
	for _, ve := range validationErrors {
		errorFields[ve.Field] = true
	}

	assert.True(t, errorFields["Username"])
	assert.True(t, errorFields["Email"])
	assert.True(t, errorFields["Password"])
	assert.True(t, errorFields["Content"])
}

func TestValidator_ValidateMaxLength(t *testing.T) {
	v := New()

	testStruct := TestStructDescription{
		Description: "This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator. This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator. This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator. This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator. This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator. This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator. This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator. This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator. This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator. This is a very long description that exceeds the maximum length allowed by the validation rules and should be rejected by the validator.", // Too long
	}

	err := v.Validate(testStruct)
	assert.Error(t, err)

	validationErrors := v.GetValidationErrors(err)
	assert.NotEmpty(t, validationErrors)

	// Vérifier que l'erreur est pour le champ Description
	hasDescriptionError := false
	for _, ve := range validationErrors {
		if ve.Field == "Description" {
			hasDescriptionError = true
			break
		}
	}
	assert.True(t, hasDescriptionError)
}
