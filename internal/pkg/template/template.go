package template

import (
	"html/template"
)

var (
	templateExt = ".html"
)

type View struct {
	Template *template.Template
	Layout   string
}

func NewEmailView(templateName string) (*template.Template, error) {
	templates := "template/email/" + templateName + ".html"

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
