package server

import (
    "embed"
    "fmt"
)

//go:embed templates
var testFS embed.FS

func TestEmbed() {
    entries, _ := testFS.ReadDir("templates")
    for _, e := range entries {
        fmt.Println(e.Name())
    }
}
