package main

import (
	"github.com/cc-jose-nieto/go-blog-gator/internal/config"
)

//var cfg config.Config = config.Config{}

func main() {
	cfg := config.Read()

	cfg.SetUser()

	config.Read()
}
