package theme_test

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Try to keep these in sync with the existing color names at theme/color.go.
var knownColorNames = []fyne.ThemeColorName{
	theme.ColorNameBackground,
	theme.ColorNameButton,
	theme.ColorNameDisabled,
	theme.ColorNameDisabledButton,
	theme.ColorNameError,
	theme.ColorNameErrorForeground,
	theme.ColorNameFocus,
	theme.ColorNameForeground,
	theme.ColorNameHeaderBackground,
	theme.ColorNameHover,
	theme.ColorNameHyperlink,
	theme.ColorNameInputBackground,
	theme.ColorNameInputBorder,
	theme.ColorNameMenuBackground,
	theme.ColorNameOverlayBackground,
	theme.ColorNamePlaceHolder,
	theme.ColorNamePressed,
	theme.ColorNamePrimary,
	theme.ColorNamePrimaryForeground,
	theme.ColorNameScrollBar,
	theme.ColorNameSelection,
	theme.ColorNameSeparator,
	theme.ColorNameShadow,
	theme.ColorNameSuccess,
	theme.ColorNameSuccessForeground,
	theme.ColorNameWarning,
	theme.ColorNameWarningForeground,
}

// Try to keep this in sync with the existing variants at theme/theme.go
var knownVariants = []fyne.ThemeVariant{
	theme.VariantDark,
	theme.VariantLight,
}

func Test_DefaultTheme_AllColorsDefined(t *testing.T) {
	th := theme.DefaultTheme()
	for _, variant := range knownVariants {
		for _, cn := range knownColorNames {
			// Transparent is used as fallback for unknown color names.
			// Built-in color names should have well-defined non-transparent values.
			assert.NotEqual(t, color.Transparent, th.Color(cn, variant), "undefined color %s variant %d", cn, variant)
		}
	}
}

func Test_DefaultTheme_PrimaryForegroundColor(t *testing.T) {
	darkColor := color.NRGBA{R: 0x17, G: 0x17, B: 0x18, A: 0xff}
	defaultTheme := theme.DefaultTheme()
	extraColorName := "some unexpected other color name where primary defaults to blue"
	lightColor := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	var testedColorNames []string
	for name, expectedColor := range map[string]color.Color{
		theme.ColorRed:    lightColor,
		theme.ColorOrange: darkColor,
		theme.ColorYellow: darkColor,
		theme.ColorGreen:  darkColor,
		theme.ColorBlue:   lightColor,
		theme.ColorPurple: lightColor,
		theme.ColorBrown:  lightColor,
		theme.ColorGray:   darkColor,
		extraColorName:    lightColor,
	} {
		if name != extraColorName {
			testedColorNames = append(testedColorNames, name)
		}
		t.Run("primary foreground color "+name, func(t *testing.T) {
			oldApp := fyne.CurrentApp()
			defer fyne.SetCurrentApp(oldApp)
			fyne.SetCurrentApp(&themedApp{theme: defaultTheme, primaryColor: name})
			t.Run("light variant", func(t *testing.T) {
				assert.Equal(t, expectedColor, defaultTheme.Color(theme.ColorNamePrimaryForeground, theme.VariantLight))
			})
			t.Run("dark variant", func(t *testing.T) {
				assert.Equal(t, expectedColor, defaultTheme.Color(theme.ColorNamePrimaryForeground, theme.VariantDark))
			})
		})
	}
	assert.ElementsMatch(t, theme.PrimaryColorNames(), testedColorNames)
}

func Test_DefaultTheme_ShadowColor(t *testing.T) {
	t.Run("light", func(t *testing.T) {
		_, _, _, a := theme.DefaultTheme().Color(theme.ColorNameShadow, theme.VariantLight).RGBA()
		assert.NotEqual(t, 255, a, "should not be transparent")
	})

	t.Run("dark", func(t *testing.T) {
		_, _, _, a := theme.DefaultTheme().Color(theme.ColorNameShadow, theme.VariantDark).RGBA()
		assert.NotEqual(t, 255, a, "should not be transparent")
	})
}

func TestEmptyTheme(t *testing.T) {
	fyne.CurrentApp().Settings().SetTheme(&emptyTheme{})
	assert.NotNil(t, theme.ForegroundColor())
	assert.NotNil(t, theme.TextFont())
	assert.NotNil(t, theme.HelpIcon())
	fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
}

func TestThemeChange(t *testing.T) {
	fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
	bg := theme.BackgroundColor()

	fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
	assert.NotEqual(t, bg, theme.BackgroundColor())
}

func TestTheme_Bootstrapping(t *testing.T) {
	current := fyne.CurrentApp().Settings().Theme()
	fyne.CurrentApp().Settings().SetTheme(nil)

	// this should not crash
	theme.BackgroundColor()

	fyne.CurrentApp().Settings().SetTheme(current)
}

type emptyTheme struct {
}

func (e *emptyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return nil
}

func (e *emptyTheme) Font(s fyne.TextStyle) fyne.Resource {
	return nil
}

func (e *emptyTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return nil
}

func (e *emptyTheme) Size(n fyne.ThemeSizeName) float32 {
	return 0
}
