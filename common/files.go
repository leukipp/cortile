package common

var (
	File FileData // Embedded file bytes
)

type FileData struct {
	Toml []byte // Default config file
	Logo []byte // Application logo icon
}

func InitFiles(toml, logo []byte) {

	// Init embedded files
	File = FileData{
		Toml: toml,
		Logo: logo,
	}
}
