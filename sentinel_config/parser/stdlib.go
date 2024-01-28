package parser

func IsStdLibName(value string) bool {
	return value == "base64" ||
		value == "decimal" ||
		value == "http" ||
		value == "json" ||
		value == "runtime" ||
		value == "sockaddr" ||
		value == "strings" ||
		value == "time" ||
		value == "types" ||
		value == "units" ||
		value == "version"
}
