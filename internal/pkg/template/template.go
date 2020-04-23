package template

import (
	"html/template"
	"path/filepath"
)

var (
	layoutDir   = "web/template/layout/"
	templateExt = ".html"
)

type View struct {
	Template *template.Template
	Layout   string
}

func NewEmailView(templateName string) (*template.Template, error) {
	templates := "web/template/email/" + templateName + ".html"

	t, err := template.New("").
		Funcs(template.FuncMap{
			"ArrToSting":         arrToSting,
			"TagsToString":       tagsToString,
			"TagsToSearchString": tagsToSearchString,
			"Add":                add,
			"Minus":              minus,
			"N":                  n,
			"IDToString":         idToString,
			"FormatTime":         formatTime,
			"ShouldDisplayTime":  shouldDisplayTime,
			"IncludesID":         includesID,
		}).
		ParseFiles(templates)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// the layout files used in our application.
func layoutFiles() []string {
	files, err := filepath.Glob(layoutDir + "*" + templateExt)
	if err != nil {
		panic(err)
	}
	return files
}
