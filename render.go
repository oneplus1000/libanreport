package libanreport

import (
	"os"
	"sync"

	"github.com/oneplus1000/errord"
	"github.com/oneplus1000/libanreport/customtextbreak"
	"github.com/signintech/pdft/render"
)

var (
	sharedThaiTextBreak   *customtextbreak.ThaiTextBreak
	sharedThaiTextBreakMu sync.RWMutex
)

func getThaiTextBreaker() (*customtextbreak.ThaiTextBreak, error) {
	sharedThaiTextBreakMu.RLock()
	if sharedThaiTextBreak != nil {
		defer sharedThaiTextBreakMu.RUnlock()
		return sharedThaiTextBreak, nil
	}
	sharedThaiTextBreakMu.RUnlock()

	sharedThaiTextBreakMu.Lock()
	defer sharedThaiTextBreakMu.Unlock()

	if sharedThaiTextBreak != nil {
		return sharedThaiTextBreak, nil
	}

	thbk := customtextbreak.NewThaiTextBreak()
	fd, err := fdLexitron.Open("customtextbreak/thaidict/lexitron.txt")
	if err != nil {
		return nil, errord.Errorf("fdLexitron.Open error: %v", err)
	}
	defer fd.Close()
	if err := thbk.LoadFromReader(fd); err != nil {
		return nil, errord.Errorf("thbk.Load error: %v", err)
	}

	sharedThaiTextBreak = thbk
	return sharedThaiTextBreak, nil
}

func bindFieldInfo(f *FieldJSON, finfo *render.FieldInfo) {
	if f.Key != nil {
		finfo.Key = *f.Key
	}
	if f.Font != nil {
		finfo.Font = *f.Font
	}
	if f.Size != nil {
		finfo.Size = *f.Size
	}
	if f.PageNum != nil {
		finfo.PageNum = *f.PageNum
	}
	if f.X != nil {
		finfo.X = convertUnit(*f.X)
	}
	if f.Y != nil {
		finfo.Y = convertUnit(*f.Y)
	}
	if f.W != nil {
		finfo.W = convertUnit(*f.W)
	}
	if f.H != nil {
		finfo.H = convertUnit(*f.H)
	}
	if f.Align != nil {
		finfo.Align = *f.Align
	}
	if f.IsWrapText != nil {
		finfo.IsWrapText = *f.IsWrapText
	}
}

func convertUnit(v float64) float64 {
	//แปลง mm ไปเป็น pdf point
	return v * 72 / 25.4
}

func newRender(tmpl Tmpl, finfos render.FieldInfos) (*render.Render, error) {
	var tmplPDFPath string
	var tmpFile *os.File

	if tmpl.embedFS != nil {
		pdfData, err := tmpl.embedFS.ReadFile(tmpl.tmplPDFPath)
		if err != nil {
			return nil, errord.Errorf("ReadFile PDF from embed error: %v", err)
		}

		tmpFile, err = os.CreateTemp("", "tmpl_*.pdf")
		if err != nil {
			return nil, errord.Errorf("CreateTemp PDF error: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.Write(pdfData)
		if err != nil {
			return nil, errord.Errorf("Write temp PDF file error: %v", err)
		}
		tmpFile.Close()

		tmplPDFPath = tmpFile.Name()

	} else {
		tmplPDFPath = tmpl.tmplPDFPath
	}

	rd, err := render.NewRender(tmplPDFPath, finfos)
	if err != nil {
		return nil, errord.Errorf("render.NewRender error: %v", err)
	}
	//rd.SetTextBreaker(textbreak.BasicTextbreak{})
	thbk, err := getThaiTextBreaker()
	if err != nil {
		return nil, err
	}
	rd.SetTextBreaker(thbk)
	return rd, nil
}
