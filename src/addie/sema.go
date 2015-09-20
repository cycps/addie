/*
This file contains the sematic checker code for Cypress designs
*/
package addie

type Diagnostic struct {
	Element string
	Level   string
	Content string
}

func Info(element, content string) Diagnostic {
	return Diagnostic{
		Element: element,
		Content: content,
		Level:   "info",
	}
}

func Warning(element, content string) Diagnostic {
	return Diagnostic{
		Element: element,
		Content: content,
		Level:   "warning",
	}
}

func Error(element, content string) Diagnostic {
	return Diagnostic{
		Element: element,
		Content: content,
		Level:   "error",
	}
}
