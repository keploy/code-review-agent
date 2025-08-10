package utils

import "strings"

// EstimateTokens provides more accurate token counting
func EstimateTokens(text string) int {
    // More sophisticated than simple /4
    // Account for code vs text differences
   // words := len(strings.Fields(text))
    chars := len(text)
    
    // Code typically has more tokens per character
    if strings.Contains(text, "func ") || strings.Contains(text, "class ") {
        return int(float64(chars) * 0.3) // Code: ~0.3 tokens per char
    }
    
    return chars / 4 // Default: 4 chars per token
}