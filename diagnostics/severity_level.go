package diagnostics

type SeverityLevel int

var Unknown SeverityLevel = 0
var Error SeverityLevel = 1
var Warning SeverityLevel = 2

func (sl SeverityLevel) ToString() string {
	switch sl {
	case Error:
		return "Error"
	case Warning:
		return "Warning"
	default:
		return "Unknown"
	}
}
