package utils

func GenerateLuhnNumber(prefix string, length int) string {
    if len(prefix) >= length {
        return prefix
    }
    digits := make([]int, length)
    for i, ch := range prefix {
        digits[i] = int(ch - '0')
    }
    for i := len(prefix); i < length-1; i++ {
        digits[i] = 0
    }
    sum := 0
    for i := length - 2; i >= 0; i-- {
        d := digits[i]
        if (length-1-i)%2 == 1 {
            d *= 2
            if d > 9 {
                d = d - 9
            }
        }
        sum += d
    }
    checkDigit := (10 - (sum % 10)) % 10
    digits[length-1] = checkDigit
    result := ""
    for _, d := range digits {
        result += string(rune(d + '0'))
    }
    return result
}

func ValidateLuhn(number string) bool {
    sum := 0
    alt := false
    for i := len(number) - 1; i >= 0; i-- {
        n := int(number[i] - '0')
        if alt {
            n *= 2
            if n > 9 {
                n = n - 9
            }
        }
        sum += n
        alt = !alt
    }
    return sum%10 == 0
}
