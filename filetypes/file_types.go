package filetypes

type FileType string

const (
	// A Configuration Override file
	ConfigOverrideFileType FileType = "override"
	// A Primary Configuration file
	ConfigPrimaryFileType FileType = "primary"
	// A Test Configuration file
	ConfigTestFileType FileType = "test"
	// A Sentinel Module file
	ModuleFileType FileType = "module"
	// A Sentinel Policy file
	PolicyFileType FileType = "policy"
)
