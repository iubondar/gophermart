package validator

import (
	"unicode"
)

// ValidateLuhn проверяет, соответствует ли строка алгоритму Луна
// https://ru.wikipedia.org/wiki/%D0%90%D0%BB%D0%B3%D0%BE%D1%80%D0%B8%D1%82%D0%BC_%D0%9B%D1%83%D0%BD%D0%B0
func ValidateLuhn(input string) bool {
	// Проверяем, что строка состоит только из цифр и не пуста
	if len(input) == 0 {
		return false
	}

	for _, r := range input {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	if len(input) == 1 {
		// Однозначные числа считаются валидными только если они равны 0
		return input[0] == '0'
	}

	sum := 0
	isSecond := false // Флаг для обработки каждой второй цифры

	// Итерируемся по строке справа налево
	for i := len(input) - 1; i >= 0; i-- {
		// Вычитая код символа '0' из кода текущего символа, мы получаем числовое значение цифры.
		digit := int(input[i] - '0')

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit = (digit / 10) + (digit % 10)
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}
