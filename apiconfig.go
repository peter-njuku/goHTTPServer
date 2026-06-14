package main

import (
	"sync/atomic"

	"github.com/peter-njuku/goHTTPServer/internal/database"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
	Db             database.Queries
	Platform       string
}
