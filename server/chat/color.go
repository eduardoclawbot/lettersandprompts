package chat

import (
	"crypto/sha256"
	"fmt"
)

// HandleColor generates a consistent color for a given handle
// Uses a hash of the handle to pick a hue, ensuring the same handle always gets the same color
func HandleColor(handle string) string {
	// Hash the handle
	hash := sha256.Sum256([]byte(handle))
	
	// Use first byte to determine hue (0-360 degrees)
	hue := int(hash[0]) * 360 / 256
	
	// Use high saturation and lightness for vibrant, readable colors
	// Format: hsl(hue, saturation%, lightness%)
	return fmt.Sprintf("hsl(%d, 70%%, 60%%)", hue)
}
