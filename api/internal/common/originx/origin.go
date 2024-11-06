package originx

import "os"

const (
	WEBUI_URL = "WEBUI_URL"
	ADMIN_URL = "ADMIN_URL"
	DEV_URL   = "DEV_URL"
)

func GetOriginDev() string {
	uri := os.Getenv(DEV_URL)
	if uri == "" {
		uri = "http://localhost:4203"
	}

	return uri
}

func GetOrigin(admin bool) string {
	if admin {
		uri := os.Getenv(ADMIN_URL)
		if uri == "" {
			uri = "http://localhost:4202"
		}
		return uri
	}

	uri := os.Getenv(WEBUI_URL)
	if uri == "" {
		uri = "http://localhost:4201"
	}

	return uri
}
