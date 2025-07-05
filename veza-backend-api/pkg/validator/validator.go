package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// Validator fournit une validation stricte des entrées avec règles de sécurité
type Validator struct {
	validate *validator.Validate
}

// New crée un nouveau validateur avec des règles de sécurité strictes
func New() *Validator {
	v := validator.New()

	// Enregistrer les validations personnalisées
	if err := v.RegisterValidation("secure_password", validateSecurePassword); err != nil {
		panic("failed to register secure_password validation: " + err.Error())
	}
	if err := v.RegisterValidation("safe_username", validateSafeUsername); err != nil {
		panic("failed to register safe_username validation: " + err.Error())
	}
	if err := v.RegisterValidation("safe_email", validateSafeEmail); err != nil {
		panic("failed to register safe_email validation: " + err.Error())
	}
	if err := v.RegisterValidation("no_sql_injection", validateNoSQLInjection); err != nil {
		panic("failed to register no_sql_injection validation: " + err.Error())
	}
	if err := v.RegisterValidation("no_xss", validateNoXSS); err != nil {
		panic("failed to register no_xss validation: " + err.Error())
	}
	if err := v.RegisterValidation("safe_filename", validateSafeFilename); err != nil {
		panic("failed to register safe_filename validation: " + err.Error())
	}
	if err := v.RegisterValidation("safe_url", validateSafeURL); err != nil {
		panic("failed to register safe_url validation: " + err.Error())
	}
	if err := v.RegisterValidation("no_special_chars", validateNoSpecialChars); err != nil {
		panic("failed to register no_special_chars validation: " + err.Error())
	}
	if err := v.RegisterValidation("max_length", validateMaxLength); err != nil {
		panic("failed to register max_length validation: " + err.Error())
	}
	if err := v.RegisterValidation("min_length", validateMinLength); err != nil {
		panic("failed to register min_length validation: " + err.Error())
	}

	return &Validator{
		validate: v,
	}
}

// Validate valide une structure avec les règles strictes
func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

// ValidateStruct alias pour Validate pour compatibilité
func (v *Validator) ValidateStruct(i interface{}) error {
	return v.Validate(i)
}

// validateSecurePassword valide un mot de passe sécurisé
func validateSecurePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Règles de sécurité strictes
	if len(password) < 8 || len(password) > 128 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateSafeUsername valide un nom d'utilisateur sécurisé
func validateSafeUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Règles strictes pour les noms d'utilisateur
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// Caractères autorisés uniquement
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(username) {
		return false
	}

	// Pas de caractères consécutifs
	if strings.Contains(username, "__") || strings.Contains(username, "--") {
		return false
	}

	return true
}

// validateSafeEmail valide un email sécurisé
func validateSafeEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	// Validation basique du format email
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(email) {
		return false
	}

	// Vérifications de sécurité - limite plus stricte pour les tests
	if len(email) > 100 { // Limite plus raisonnable pour les applications
		return false
	}

	// Vérifier la longueur du domaine
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[1]) > 63 { // Limite plus stricte pour les domaines
		return false
	}

	// Pas de caractères dangereux
	dangerousChars := []string{"<", ">", "\"", "'", "&", "script", "javascript"}
	for _, char := range dangerousChars {
		if strings.Contains(strings.ToLower(email), char) {
			return false
		}
	}

	return true
}

// validateNoSQLInjection détecte les tentatives d'injection SQL
func validateNoSQLInjection(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())

	// Mots-clés SQL dangereux
	sqlKeywords := []string{
		"select", "insert", "update", "delete", "drop", "create", "alter",
		"union", "exec", "execute", "declare", "cast", "convert",
		"waitfor", "delay", "xp_", "sp_", "0x", "--", "/*", "*/",
	}

	for _, keyword := range sqlKeywords {
		if strings.Contains(value, keyword) {
			return false
		}
	}

	return true
}

// validateNoXSS détecte les tentatives d'injection XSS
func validateNoXSS(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())

	// Patterns XSS dangereux
	xssPatterns := []string{
		"<script", "javascript:", "vbscript:", "onload=", "onerror=",
		"onclick=", "onmouseover=", "onfocus=", "onblur=",
		"<iframe", "<object", "<embed", "<applet", "<meta",
		"expression(", "eval(", "alert(", "confirm(", "prompt(",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(value, pattern) {
			return false
		}
	}

	return true
}

// validateSafeFilename valide un nom de fichier sécurisé
func validateSafeFilename(fl validator.FieldLevel) bool {
	filename := fl.Field().String()

	// Caractères interdits dans les noms de fichiers
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return false
		}
	}

	// Pas de noms de fichiers système (sans extension)
	systemFiles := []string{"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}

	// Extraire le nom de fichier sans extension
	nameWithoutExt := filename
	if strings.Contains(filename, ".") {
		parts := strings.Split(filename, ".")
		nameWithoutExt = parts[0]
	}

	for _, sysFile := range systemFiles {
		if strings.ToLower(nameWithoutExt) == sysFile {
			return false
		}
	}

	// Vérifier l'extension du nom de fichier
	if strings.Contains(filename, ".") {
		parts := strings.Split(filename, ".")
		if len(parts) > 1 {
			ext := strings.ToLower(parts[len(parts)-1])
			// Extensions dangereuses
			dangerousExts := []string{"exe", "bat", "cmd", "com", "pif", "scr", "vbs", "js", "jar", "msi"}
			for _, dangerousExt := range dangerousExts {
				if ext == dangerousExt {
					return false
				}
			}
		}
	}

	return true
}

// validateSafeURL valide une URL sécurisée
func validateSafeURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()

	// Protocoles autorisés uniquement
	allowedProtocols := []string{"http://", "https://", "ftp://", "sftp://"}
	hasValidProtocol := false

	for _, protocol := range allowedProtocols {
		if strings.HasPrefix(strings.ToLower(url), protocol) {
			hasValidProtocol = true
			break
		}
	}

	if !hasValidProtocol {
		return false
	}

	// Pas de caractères dangereux
	dangerousChars := []string{"<", ">", "\"", "'", "javascript:", "vbscript:"}
	for _, char := range dangerousChars {
		if strings.Contains(strings.ToLower(url), char) {
			return false
		}
	}

	return true
}

// validateNoSpecialChars valide qu'il n'y a pas de caractères spéciaux dangereux
func validateNoSpecialChars(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Caractères spéciaux dangereux
	dangerousChars := []rune{'<', '>', '"', '\'', '&', ';', '|', '`', '$', '(', ')', '{', '}', '[', ']'}

	for _, char := range value {
		for _, dangerous := range dangerousChars {
			if char == dangerous {
				return false
			}
		}
	}

	return true
}

// validateMaxLength valide la longueur maximale
func validateMaxLength(fl validator.FieldLevel) bool {
	maxLength := fl.Param()
	value := fl.Field().String()

	// Convertir le paramètre en int
	var max int
	if _, err := fmt.Sscanf(maxLength, "%d", &max); err != nil {
		return false
	}

	return len(value) <= max
}

// validateMinLength valide la longueur minimale
func validateMinLength(fl validator.FieldLevel) bool {
	minLength := fl.Param()
	value := fl.Field().String()

	// Convertir le paramètre en int
	var min int
	if _, err := fmt.Sscanf(minLength, "%d", &min); err != nil {
		return false
	}

	return len(value) >= min
}

// ValidationError représente une erreur de validation avec détails
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// GetValidationErrors retourne les erreurs de validation détaillées
func (v *Validator) GetValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErr {
			validationErrors = append(validationErrors, ValidationError{
				Field:   fieldErr.Field(),
				Tag:     fieldErr.Tag(),
				Value:   fmt.Sprintf("%v", fieldErr.Value()),
				Message: getErrorMessage(fieldErr),
			})
		}
	}

	return validationErrors
}

// getErrorMessage retourne un message d'erreur lisible
func getErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fmt.Sprintf("Le champ %s est requis", fieldErr.Field())
	case "email":
		return fmt.Sprintf("Le champ %s doit être un email valide", fieldErr.Field())
	case "secure_password":
		return "Le mot de passe doit contenir au moins 8 caractères, une majuscule, une minuscule, un chiffre et un caractère spécial"
	case "safe_username":
		return "Le nom d'utilisateur doit contenir entre 3 et 30 caractères alphanumériques, tirets ou underscores"
	case "no_sql_injection":
		return fmt.Sprintf("Le champ %s contient des caractères non autorisés", fieldErr.Field())
	case "no_xss":
		return fmt.Sprintf("Le champ %s contient du contenu non autorisé", fieldErr.Field())
	case "safe_filename":
		return fmt.Sprintf("Le nom de fichier contient des caractères non autorisés")
	case "safe_url":
		return fmt.Sprintf("L'URL doit utiliser un protocole sécurisé (http, https, ftp, sftp)")
	case "max_length":
		return fmt.Sprintf("Le champ %s ne doit pas dépasser %s caractères", fieldErr.Field(), fieldErr.Param())
	case "min_length":
		return fmt.Sprintf("Le champ %s doit contenir au moins %s caractères", fieldErr.Field(), fieldErr.Param())
	default:
		return fmt.Sprintf("Le champ %s n'est pas valide", fieldErr.Field())
	}
}
