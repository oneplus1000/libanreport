package libanreport

import (
	"embed"
)

// TmplJSON tmpl.json
type TmplJSON struct {
	Fonts  []FontJSON
	Fields []FieldJSON
	Styles []StyleJSON
}

func (t TmplJSON) stylesIndexByName(name string) int {
	for i, s := range t.Styles {
		if s.Name == name {
			return i
		}
	}
	return -1
}

// FieldJSON field
type FieldJSON struct {
	Key        *string
	Style      *string
	Font       *string
	Size       *int
	PageNum    *int
	X, Y       *float64
	W, H       *float64
	Align      *int
	IsWrapText *bool
}

// FontJSON font
type FontJSON struct {
	Name string
	TTF  string
}

// StyleJSON style
type StyleJSON struct {
	Name   string
	Define FieldJSON
}

// DefineStyleJSON define style
type DefineStyleJSON struct {
	FieldJSON
}

// FontOverrideJSON fonts.json filed
type FontOverrideJSON struct {
	Name      string
	Textrises []TextriseJSON
	Kernings  []KerningJSON
}

// TextriseJSON text rise override
type TextriseJSON struct {
	Lefts  []string
	Rights []string
	Nexts  []string
	Val    float32
}

// KerningJSON kerning override
type KerningJSON struct {
	Lefts  []string
	Rights []string
	Val    int16
}

// Tmpl tmpl
type Tmpl struct {
	tmplFolderPath string
	folderName     string
	tmplPDFPath    string
	tmplJSON       *TmplJSON
	embedFS        *embed.FS
	embedPath      string
}
