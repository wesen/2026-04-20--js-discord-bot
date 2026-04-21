// Package doc embeds the Discord bot help pages.
package doc

import (
	"embed"

	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed topics/*.md tutorials/*.md
var docFS embed.FS

// AddDocToHelpSystem loads the embedded help sections into the application help system.
func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
