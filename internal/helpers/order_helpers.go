package helpers

func convertStringToIntSlice(str string) ([]int, error) {
	nums := make([]int, len(str))
	for i, s := range str {
		nums[i] = int(s - '0')
	}

	return nums, nil
}

func LuhnCheck(orderNum string) (bool, error) {
	orderNums, err := convertStringToIntSlice(orderNum)
	if err != nil {
		return false, err
	}

	if len(orderNums) == 0 {
		return false, nil
	}

	sum := 0
	parity := len(orderNums) % 2

	for i := range orderNums {
		digit := orderNums[i]
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0, nil
}
