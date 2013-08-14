package crocodoc

var Http_4xx_errors map[int]string

func init() {
	if CROCODOC_API_TOKEN == "YOUR_TOKEN_HERE" {
		panic("Please set your token in config.go. Exiting.")
	}
	Http_4xx_errors = map[int]string{
		400: "Bad Request",
		401: "Unauthorized",
		404: "Not found",
		405: "Method not allowed",
	}
}
