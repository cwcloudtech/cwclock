package handlers

import (
	"cwclock/config"
	"fmt"
)

func HandlerGetApiURL() {
	apiURL := config.GetApiURL()
	fmt.Printf("API URL = %v\n", apiURL)
}

func HandlerSetApiURL(value string) {
	config.SetApiURL(value)
	fmt.Printf("API URL = %v\n", value)
}

func HandlerSetDefaultFormat(value string) {
	config.SetDefaultFormat(value)
	fmt.Printf("Output format = %v\n", value)
}
