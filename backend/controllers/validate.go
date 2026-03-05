package controllers

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// customEmojiIDRe matches a bare snowflake numeric ID used for custom emoji.
var customEmojiIDRe = regexp.MustCompile(`^\d{15,20}$`)

// emojiBaseASCII are the printable ASCII code-points that may appear inside a
// valid emoji sequence (keycap base characters: #, *, 0-9).
var emojiBaseASCII = map[rune]bool{
	'#': true, '*': true,
	'0': true, '1': true, '2': true, '3': true, '4': true,
	'5': true, '6': true, '7': true, '8': true, '9': true,
}

// isValidEmoji reports whether s is a valid emoji value.
// It accepts Unicode emoji sequences (including modifiers, ZWJ, and variation
// selectors) and bare numeric custom-emoji snowflake IDs (15–20 digits).
// Plain ASCII text (letters, punctuation, etc.) is rejected.
func isValidEmoji(s string) bool {
	if customEmojiIDRe.MatchString(s) {
		return true
	}
	hasNonASCII := false
	for _, r := range s {
		switch {
		case r > 0x7F:
			// Non-ASCII: emoji codepoint, modifier, ZWJ (U+200D), variation selector, etc.
			hasNonASCII = true
		case emojiBaseASCII[r]:
			// Keycap-base ASCII allowed when part of a composite emoji (e.g. 1️⃣)
		default:
			return false
		}
	}
	return hasNonASCII // must include at least one non-ASCII rune
}

// validateBody validates a struct using its `validate` tags.
// Returns a human-readable error message, or "" if valid.
func validateBody(s interface{}) string {
	if err := validate.Struct(s); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
			f := ve[0]
			switch f.Tag() {
			case "required":
				return f.Field() + " is required"
			case "max":
				return f.Field() + " is too long (max " + f.Param() + " characters)"
			case "min":
				return f.Field() + " is too short (min " + f.Param() + " characters)"
			case "oneof":
				return f.Field() + " must be one of: " + f.Param()
			}
			return f.Field() + " is invalid"
		}
		return "Invalid input"
	}
	return ""
}
