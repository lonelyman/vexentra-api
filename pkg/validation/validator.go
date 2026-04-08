// vexentra-api/pkg/validation/validator.go
package validation

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// ====================================================================================
// Structs for Validation Result
// ====================================================================================

// ValidationResult holds the complete validation result.
type ValidationResult struct {
	IsValid bool                    `json:"is_valid"`
	Errors  []ValidationErrorDetail `json:"errors,omitempty"`
}

// ValidationErrorDetail represents a single validation error.
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

// ====================================================================================
// Validator Factory & Custom Rules
// ====================================================================================

// New creates and configures a new validator instance with all project-wide custom rules.
func New() *validator.Validate {
	v := validator.New()

	// strong_password: min 8 chars, at least 1 upper, 1 lower, 1 digit, 1 special character
	_ = v.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		var hasUpper, hasLower, hasDigit, hasSpecial bool
		val := fl.Field().String()
		if len(val) < 8 {
			return false
		}
		for _, ch := range val {
			switch {
			case unicode.IsUpper(ch):
				hasUpper = true
			case unicode.IsLower(ch):
				hasLower = true
			case unicode.IsDigit(ch):
				hasDigit = true
			case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
				hasSpecial = true
			}
		}
		return hasUpper && hasLower && hasDigit && hasSpecial
	})

	return v
}

// ====================================================================================
// Main Validation Function
// ====================================================================================

// Validate validates a struct and returns a structured, detailed result.
func Validate(v *validator.Validate, s interface{}) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationErrorDetail{},
	}

	err := v.Struct(s)
	if err != nil {
		result.IsValid = false
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			result.Errors = translateValidationErrors(s, validationErrors)
		}
		return result
	}

	return result
}

// ====================================================================================
// Error Translation & Parsing Helpers
// ====================================================================================

// translateValidationErrors converts validator.ValidationErrors to our custom format.
func translateValidationErrors(s interface{}, validationErrors validator.ValidationErrors) []ValidationErrorDetail {
	var details []ValidationErrorDetail
	structType := reflect.TypeOf(s)
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	for _, ve := range validationErrors {
		var fieldName string
		var customMessage string

		if field, ok := structType.FieldByName(ve.Field()); ok {
			// Get field name from json tag
			jsonTag := field.Tag.Get("json")
			fieldName = strings.Split(jsonTag, ",")[0]
			if fieldName == "" {
				fieldName = ve.Field()
			}

			// Parse the vmsg tag to find a specific message for the failed rule
			vmsgTag := field.Tag.Get("vmsg")
			messageMap := parseVmsgTag(vmsgTag)
			customMessage = messageMap[ve.Tag()]
		} else {
			fieldName = ve.Field()
		}

		details = append(details, ValidationErrorDetail{
			Field:   fieldName,
			Message: ternary(customMessage != "", customMessage, generateDefaultMessage(ve.Tag(), ve.Param(), ve)),
			Value:   fmt.Sprintf("%v", ve.Value()),
		})
	}
	return details
}

// parseVmsgTag parses the vmsg struct tag into a map of rule -> message.
// Supports escaped commas via \, in the tag value.
func parseVmsgTag(tag string) map[string]string {
	messageMap := make(map[string]string)
	if tag == "" {
		return messageMap
	}

	var parts []string
	var current strings.Builder
	escaped := false
	for _, char := range tag {
		if char == '\\' && !escaped {
			escaped = true
			continue
		}
		if char == ',' && !escaped {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(char)
		}
		escaped = false
	}
	parts = append(parts, current.String())

	for _, rule := range parts {
		keyValue := strings.SplitN(rule, ":", 2)
		if len(keyValue) == 2 {
			ruleName := strings.TrimSpace(keyValue[0])
			message := strings.ReplaceAll(strings.TrimSpace(keyValue[1]), `\,`, `,`)
			messageMap[ruleName] = message
		}
	}
	return messageMap
}

// generateDefaultMessage creates a Thai user-friendly error message.
// Falls back to the original library error message for unknown tags.
func generateDefaultMessage(tag, param string, originalError error) string {
	switch tag {
	// General
	case "required":
		return "ฟิลด์นี้จำเป็นต้องระบุ"
	case "email":
		return "ต้องเป็นรูปแบบอีเมลที่ถูกต้อง"
	case "url":
		return "ต้องเป็น URL ที่ถูกต้อง"
	case "uuid":
		return "ต้องเป็น UUID ที่ถูกต้อง"

	// Length / Size
	case "min":
		return fmt.Sprintf("ต้องมีขนาดอย่างน้อย %s", param)
	case "max":
		return fmt.Sprintf("ต้องมีขนาดไม่เกิน %s", param)
	case "len":
		return fmt.Sprintf("ต้องมีขนาดเท่ากับ %s พอดี", param)

	// Numeric
	case "numeric":
		return "ต้องเป็นตัวเลขเท่านั้น"
	case "gt":
		return fmt.Sprintf("ต้องมีค่ามากกว่า %s", param)
	case "gte":
		return fmt.Sprintf("ต้องมีค่าอย่างน้อย %s", param)
	case "lt":
		return fmt.Sprintf("ต้องมีค่าน้อยกว่า %s", param)
	case "lte":
		return fmt.Sprintf("ต้องมีค่าไม่เกิน %s", param)

	// Comparison
	case "eq":
		return fmt.Sprintf("ต้องมีค่าเท่ากับ %s", param)
	case "ne":
		return fmt.Sprintf("ต้องมีค่าไม่เท่ากับ %s", param)
	case "eqfield":
		return fmt.Sprintf("ค่าต้องตรงกับฟิลด์ %s", param)

	// String format
	case "alphanum":
		return "ต้องเป็นตัวอักษรหรือตัวเลขเท่านั้น"
	case "alpha":
		return "ต้องเป็นตัวอักษรเท่านั้น"

	// Custom rules
	case "strong_password":
		return "รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร ประกอบด้วยตัวพิมพ์ใหญ่ พิมพ์เล็ก ตัวเลข และอักขระพิเศษ"

	default:
		return originalError.Error()
	}
}

func ternary(condition bool, ifTrue, ifFalse string) string {
	if condition {
		return ifTrue
	}
	return ifFalse
}
