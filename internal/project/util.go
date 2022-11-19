package project

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// devMode indicates whether the project is running in development mode.
var devMode, _ = strconv.ParseBool(os.Getenv("DEV_MODE"))

// DevMode indicates whether the project is running in development mode.
func DevMode() bool {
	return devMode
}

// root is the path to this file parent module.
var root = func() string {
	_, self, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(self), "..", "..")
}()

// Root returns the filepath to the root of this project.
func Root(more ...string) string {
	if len(more) == 0 {
		return root
	}

	parts := make([]string, 0, len(more)+1)
	parts = append(parts, root)
	parts = append(parts, more...)
	return filepath.Join(parts...)
}
