package crocodoc

var (
	CrocoDocToken string

	Http4xxErrors map[int]string = map[int]string{
		400: "Bad Request",
		401: "Unauthorized",
		404: "Not found",
		405: "Method not allowed",
	}
)
