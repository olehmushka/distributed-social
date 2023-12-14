package stringutils

import "strconv"

// StringToInt converts input string to output integer.
func StringToInt(in string) (int, error) {
	out, err := strconv.Atoi(in)
	if err != nil {
		return 0, err
	}

	return out, nil
}

// StringToInt64 converts input string to output integer 64.
func StringToInt64(in string) (int64, error) {
	out, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		return 0, err
	}

	return out, nil
}

// StringToInt32 converts input string to output integer 32.
func StringToInt32(in string) (int32, error) {
	out, err := strconv.ParseInt(in, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(out), nil
}

// StringToUInt64 converts input string to output unsigned integer 64.
func StringToUInt64(in string) (uint64, error) {
	out, err := strconv.ParseUint(in, 10, 64)
	if err != nil {
		return 0, err
	}

	return out, nil
}

// StringToUInt32 converts input string to output unsigned integer 32.
func StringToUInt32(in string) (uint32, error) {
	out, err := strconv.ParseUint(in, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(out), nil
}

// StringToUInt16 converts input string to output unsigned integer 16.
func StringToUInt16(in string) (uint16, error) {
	out, err := strconv.ParseUint(in, 10, 16)
	if err != nil {
		return 0, err
	}

	return uint16(out), nil
}

// StringToFloat64 converts input string to output float64.
func StringToFloat64(in string) (float64, error) {
	out, err := strconv.ParseFloat(in, 64)
	if err != nil {
		return 0, err
	}

	return out, nil
}

// StringToFloat32 converts input string to output float32.
func StringToFloat32(in string) (float32, error) {
	out, err := strconv.ParseFloat(in, 32)
	if err != nil {
		return 0, err
	}

	return float32(out), nil
}

// StringToBool converts input string to output boolean.
func StringToBool(in string) (bool, error) {
	out, err := strconv.ParseBool(in)
	if err != nil {
		return false, err
	}

	return out, nil
}
