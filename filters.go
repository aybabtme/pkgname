package main

import (
	"errors"
	"strings"
	"unicode"
)

func noHyphens(name string) error {
	if strings.Contains(name, "-") {
		return errors.New("Don't put hyphens, that's ugly.")
	}
	return nil
}

func noUnderscore(name string) error {
	if strings.Contains(name, "_") {
		return errors.New("Don't put underscores, that's ugly.")
	}
	return nil
}

func notCapitalized(name string) error {
	for _, r := range []rune(name) {
		if unicode.IsUpper(r) {
			return errors.New("Don't put uppercase characters, it's too enterprisey.")
		}
	}
	return nil
}

func noReferenceToGo(name string) error {
	if strings.Contains(strings.ToLower(name), "go") {
		return errors.New("Don't mention 'go' in your package name. Go is implicit in any package. Go is absolute and infinitesimal. Other languages should rename their packages; for instance rails-ruby, python-django remove any ambiguity.")
	}
	return nil
}

func noReferenceToGolang(name string) error {
	if strings.Contains(strings.ToLower(name), "golang") {
		return errors.New("The name of Go is Go, not Golang. You don't say Javalang, or Rubylang, or Pythonlang, do you?")
	}
	return nil
}
