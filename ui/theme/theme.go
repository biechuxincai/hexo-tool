//go:generate fyne bundle -o bundled.go GochiHand.ttf
//go:generate fyne bundle -append -o bundled.go Icon.png

package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MyTheme struct {
}

func (m *MyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		if v == theme.VariantLight {
			return &color.NRGBA{R: 213, G: 245, B: 213, A: 0xFF}
		}
		return &color.NRGBA{R: 0x37, G: 0x2B, B: 0x09, A: 0xFF}
	case theme.ColorNameForeground:
		if v == theme.VariantLight {
			return &color.NRGBA{R: 0x46, G: 0x3A, B: 0x11, A: 0xFF}
		}
		return &color.NRGBA{R: 0xF0, G: 0xE9, B: 0x9B, A: 0xFF}
	//case theme.ColorNamePrimary:
	//	return &color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xAA}
	//case theme.ColorNameFocus:
	//	return &color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x66}
	case theme.ColorNameInputBackground:
		// removes background from entry
		return color.Transparent
		//return &color.RGBA{R: 198, G: 210, B: 16, A: 75}
	}

	return theme.DefaultTheme().Color(n, v)
}

func (m *MyTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTextFont()
	//return resourceGochiHandTtf
}

func (m *MyTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (m *MyTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameText {
		return theme.DefaultTheme().Size(n) + 2
	}

	return theme.DefaultTheme().Size(n)
}
