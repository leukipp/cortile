package common

var File FileData

type FileData struct {
	Toml []byte // Default config file
	Icon []byte // Application icon
}

func InitFiles(toml, icon []byte) {

	// Init embedded files
	File = FileData{
		Toml: toml,
		Icon: icon,
	}
}
