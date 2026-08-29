package docs

import (
	"embed"
	"io/fs"
)

//go:embed content/en
var skills embed.FS

var Skills, _ = fs.Sub(skills, "content/en")

//go:embed SKILL.md
var SkillMd embed.FS
