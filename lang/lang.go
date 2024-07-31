// Package lang introduces a translation and localisation API for Fyne applications
//
// Since 2.5
package lang

import (
	"embed"
	"encoding/json"
	"log"
	"sort"
	"strings"
	"text/template"

	"github.com/jeandeaual/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"fyne.io/fyne/v2"

	"golang.org/x/text/language"
)

var (
	// L is a shortcut to localize a string, similar to the gettext "_" function.
	// More info available on the `Localize` function.
	L = Localize

	// N is a shortcut to localize a string with plural forms, similar to the ngettext function.
	// More info available on the `LocalizePlural` function.
	N = LocalizePlural

	// X is a shortcut to get the localization of a string with specified key, similar to pgettext.
	// More info available on the `LocalizeKey` function.
	X = LocalizeKey

	// XN is a shortcut to get the localization plural form of a string with specified key, similar to npgettext.
	// More info available on the `LocalizePluralKey` function.
	XN = LocalizePluralKey

	// This defines an order in which it will try to find a fallback in case localizer does not find a match.
	// All other languages will be in order as the system reads them (which is most likely alphabetical).
	languageOrder = []string{"en"}

	preferredLanguage string

	bundle    *i18n.Bundle
	localizer *i18n.Localizer

	//go:embed translations
	translations embed.FS
	translated   []language.Tag
)

// Localize asks the translation engine to translate a string, this behaves like the gettext "_" function.
// The string can be templated and the template data can be passed as a struct with exported fields,
// or as a map of string keys to any suitable value.
func Localize(in string, data ...any) string {
	return LocalizeKey(in, in, data...)
}

// LocalizeKey asks the translation engine for the translation with specific ID.
// If it cannot be found then the fallback will be used.
// The string can be templated and the template data can be passed as a struct with exported fields,
// or as a map of string keys to any suitable value.
func LocalizeKey(key, fallback string, data ...any) string {
	var d0 any
	if len(data) > 0 {
		d0 = data[0]
	}

	ret, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    key,
			Other: fallback,
		},
		TemplateData: d0,
	})

	if err != nil {
		fyne.LogError("Translation failure", err)
		return fallbackWithData(key, fallback, d0)
	}
	return ret
}

// LocalizePlural asks the translation engine to translate a string from one of a number of plural forms.
// This behaves like the ngettext function, with the `count` parameter determining the plurality looked up.
// The string can be templated and the template data can be passed as a struct with exported fields,
// or as a map of string keys to any suitable value.
func LocalizePlural(in string, count int, data ...any) string {
	return LocalizePluralKey(in, in, count, data...)
}

// LocalizePluralKey asks the translation engine for the translation with specific ID in plural form.
// This behaves like the npgettext function, with the `count` parameter determining the plurality looked up.
// If it cannot be found then the fallback will be used.
// The string can be templated and the template data can be passed as a struct with exported fields,
// or as a map of string keys to any suitable value.
func LocalizePluralKey(key, fallback string, count int, data ...any) string {
	var d0 any
	if len(data) > 0 {
		d0 = data[0]
	}

	ret, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    key,
			Other: fallback,
		},
		PluralCount:  count,
		TemplateData: d0,
	})

	if err != nil {
		fyne.LogError("Translation failure", err)
		return fallbackWithData(key, fallback, d0)
	}
	return ret
}

// AddTranslations allows an app to load a bundle of translations.
// The language that this relates to will be inferred from the resource name, for example "fr.json".
// The data should be in json format.
func AddTranslations(r fyne.Resource) error {
	defer updateLocalizer()
	return addLanguage(r.Content(), r.Name())
}

// AddTranslationsForLocale allows an app to load a bundle of translations for a specified locale.
// The data should be in json format.
func AddTranslationsForLocale(data []byte, l fyne.Locale) error {
	defer updateLocalizer()
	return addLanguage(data, l.String()+".json")
}

// AddTranslationsFS supports adding all translations in one calling using an `embed.FS` setup.
// The `dir` parameter specifies the name or path of the directory containing translation files
// inside this embedded filesystem.
// Each file should be a json file with the name following pattern [prefix.]lang.json.
func AddTranslationsFS(fs embed.FS, dir string) (retErr error) {
	files, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		name := f.Name()
		data, err := fs.ReadFile(dir + "/" + name)
		if err != nil {
			if retErr == nil {
				retErr = err
			}
			continue
		}

		err = addLanguage(data, name)
		if err != nil {
			if retErr == nil {
				retErr = err
			}
			continue
		}
	}

	updateLocalizer()

	return retErr
}

// SetLanguageOrder allows an app to set the order in which translations are checked in case no locale matches.
// Since 2.6
func SetLanguageOrder(order []string) {
	languageOrder = order
	updateLocalizer()
}


// SetPreferredLocale allows an app to set the preferred locale for translations, overwriting the System Locale.
// locale can be in format en_US_someVariant, en_US, en-US-someVariant, en-US, en
// Since 2.6
func SetPreferredLocale(locale string) {
	preferredLanguage = locale
	updateLocalizer()
}

func addLanguage(data []byte, name string) error {
	f, err := bundle.ParseMessageFileBytes(data, name)
	translated = append(translated, f.Tag)
	return err
}

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	err := AddTranslationsFS(translations, "translations")
	if err != nil {
		fyne.LogError("Error occurred loading built-in translations", err)
	}
}

func fallbackWithData(key, fallback string, data any) string {
	t, err := template.New(key).Parse(fallback)
	if err != nil {
		log.Println("Could not parse fallback template")
		return fallback
	}
	str := &strings.Builder{}
	_ = t.Execute(str, data)
	return str.String()
}

func orderLanguages(a, b language.Tag) bool {
	indexA := -1
	indexB := -1
	for i, l := range languageOrder {
		if a.String() == l {
			indexA = i
		}
		if b.String() == l {
			indexB = i
		}
	}
	// Order both languages as defined in languageOrder
	if indexA != -1 && indexB != -1 {
		return indexA < indexB
	}
	// If it is the only language in languageOrder, it comes first
	if indexA != -1 {
		return true
	}
	if indexB != -1 {
		return false
	}
	// If no language is in languageOrder, sort alphabetically
	return strings.Compare(a.String(), b.String()) < 0
}

// A utility for setting up languages - available to unit tests for overriding system
func setupLang(lang string) {
	localizer = i18n.NewLocalizer(bundle, lang)
}

// updateLocalizer Finds the closest translation from the user's locale list and sets it up
func updateLocalizer() {
	// Sort the translated slice using the orderLanguages function
	sort.SliceStable(translated, func(i, j int) bool {
		return orderLanguages(translated[i], translated[j])
	})

	all, err := locale.GetLocales()
	if err != nil {
		fyne.LogError("Failed to load user locales", err)
		all = []string{"en"}
	}
	if preferredLanguage != "" {
		all = []string{preferredLanguage}
	}
	str := closestSupportedLocale(all).LanguageString()
	setupLang(str)
	localizer = i18n.NewLocalizer(bundle, str)
}
