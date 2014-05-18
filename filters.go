package main

import (
	"errors"
	"fmt"
	"github.com/grd/stat"
	"log"
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
		return errors.New("Don't mention 'go' in your package name. Go is implicit in any package. Go is absolute and infinitesimal. Other languages should rename their packages; for instance 'rails-ruby' and 'python-django' would remove any ambiguity.")
	}
	return nil
}

func noReferenceToGolang(name string) error {
	if strings.Contains(strings.ToLower(name), "golang") {
		return errors.New("The name of Go is Go, not Golang. You don't say Javalang, or Rubylang, or Pythonlang, do you?")
	}
	return nil
}

var errInvalidPackage = "That's not even a valid package name: %s!" +
	" Read the spec: http://golang.org/ref/spec#Package_clause"

func validPackageNames(name string) error {
	if len(name) < 1 {
		return fmt.Errorf(errInvalidPackage, "the name can't be blank")
	}

	for i, r := range []rune(name) {
		if i == 0 {
			if !unicode.IsLetter(r) {
				return fmt.Errorf(errInvalidPackage, "the first character must be a letter")
			}
		}

		switch {
		case unicode.IsLetter(r):
		case unicode.IsDigit(r):
		case r == '-':
		case r == '_':
		case r == '.':
			// ok
		default:
			return fmt.Errorf(errInvalidPackage, "all the characters (but the first) must be either letters or digits")
		}
	}

	return nil
}

func closeToMean(allnames []string, maxDist float64) (f Filter, mean, stdev float64) {
	data := make(stat.IntSlice, len(allnames))
	for i, name := range allnames {
		data[i] = int64(len(name))
	}

	mean = stat.Mean(data)
	stdev = stat.SdMean(data, mean)

	maxMean := int(mean + stdev*maxDist)

	f = func(name string) error {
		diff := float64(len(name)) - mean
		dist := diff / stdev

		log.Printf("len=%d\tdiff=%f\tdist=%f\tmean=%f\tsd=%f", len(name), diff, dist, mean, stdev)

		if dist > maxDist {
			return fmt.Errorf("This package name is %.1f std.dev. longer than normal."+
				" It should at most %d characters long.", dist, maxMean)
		}
		return nil
	}
	return
}
