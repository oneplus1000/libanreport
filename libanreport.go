package libanreport

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/oneplus1000/errord"
	"github.com/signintech/pdft/render"
)

//go:embed "customtextbreak/thaidict/lexitron.txt"
var fdLexitron embed.FS

// TypeText text
const TypeText = 1

// TypeIMG img
const TypeIMG = 2

var ErrFileNotFound = errors.New("file not found")

// สร้าง PDF จากเทมเพลตและข้อมูล
func GenPdf(
	tmpl Tmpl,
	datas []DataJSON,
	fontoverrides []FontOverrideJSON,
	w io.Writer,
) error {

	var finfos render.FieldInfos
	for _, f := range tmpl.tmplJSON.Fields {
		var finfo render.FieldInfo
		var styleDefine *FieldJSON
		if f.Style != nil {
			index := tmpl.tmplJSON.stylesIndexByName(*f.Style)
			if index < 0 {
				return errors.New("Fields.Style not found")
			}
			styleDefine = &tmpl.tmplJSON.Styles[index].Define
		}

		if styleDefine != nil {
			bindFieldInfo(styleDefine, &finfo)
		}
		bindFieldInfo(&f, &finfo)

		//get IsWrapText from datas (if exists)
		datasIdx, ok := findDataJSONIndexByKey(datas, *f.Key)
		if ok {
			if datas[datasIdx].WrapTextType == WrapTextTypeNewLine {
				finfo.IsWrapText = true
			}
		}

		finfos = append(finfos, finfo)
	}

	rdr, err := newRender(tmpl, finfos)
	if err != nil {
		return errord.Errorf("newRender error: %w", err)
	}

	//add font
	for _, f := range tmpl.tmplJSON.Fonts {
		if tmpl.embedFS != nil {
			fontpath := filepath.Join(tmpl.embedPath, f.TTF)
			fontData, err := tmpl.embedFS.ReadFile(fontpath)
			if err != nil {
				return errord.Errorf("ReadFile from embed error: %w", err)
			}

			tmpFile, err := os.CreateTemp("", f.Name+"_*.ttf")
			if err != nil {
				return errord.Errorf("CreateTemp error: %w", err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			_, err = tmpFile.Write(fontData)
			if err != nil {
				return errord.Errorf("Write temp font file error: %w", err)
			}
			tmpFile.Close()

			err = rdr.AddFont(f.Name, tmpFile.Name())
			if err != nil {
				return errord.Errorf("AddFont error: %w", err)
			}
		} else {
			fontpath := filepath.Join(tmpl.tmplFolderPath, f.TTF)
			err = rdr.AddFont(f.Name, fontpath)
			if err != nil {
				return errord.Errorf("AddFont error: %w", err)
			}
		}
	}

	//start setup override
	for _, fontoverride := range fontoverrides {
		rdr.TextriseOverride(fontoverride.Name, func(
			leftRune rune,
			rightRune rune,
			fontSize int,
			allText string,
			currTextIndex int,
		) float32 {
			gap := float32(0)
			runes := []rune(allText)
			var nextRune rune
			if len(runes) > currTextIndex+1 {
				nextRune = runes[currTextIndex+1]
			}
			for _, textrise := range fontoverride.Textrises {
				if isInRune(leftRune, textrise.Lefts) && isInRune(rightRune, textrise.Rights) && isInRune(nextRune, textrise.Nexts) {
					gap = textrise.Val
					break
				}
			}
			return gap * float32(fontSize) / 100
		})

		k := int64(0)
		lastK := int64(0)
		lastPairVal := int16(0)
		rdr.KernOverride(fontoverride.Name, func(
			leftRune rune,
			rightRune rune,
			leftPair uint,
			rightPair uint,
			pairVal int16,
		) int16 {
			for _, kern := range fontoverride.Kernings {
				if isInRune(leftRune, kern.Lefts) && isInRune(rightRune, kern.Rights) {
					pairVal = kern.Val
					lastPairVal = kern.Val
					lastK = k
				}
			}
			if k-1 == lastK { //ทำเพื่อโยกตัวหลังจากทีี่ถูกปรับ  kerning ให้ไปวางที่เดิมไม่ถูกบีบ
				pairVal = (-1) * lastPairVal
			}
			k++
			return pairVal
		})
	}
	//end setup override

	//bind data
	for _, d := range datas {
		if d.Type == TypeText {
			err = rdr.Text(d.Key, d.Val)
			if err != nil {
				return errord.Errorf("Error on key %s: %s", d.Key, err)
			}
		} else if d.Type == TypeIMG {
			err = rdr.ImgBase64(d.Key, d.Val)
			if err != nil {
				return errord.Errorf("Error on key %s: %s", d.Key, err)
			}
		}
	}

	//save pdf
	err = rdr.SaveTo(w)
	if err != nil {
		return errord.Errorf("rd.SaveTo error: %v", err)
	}

	return nil
}

func findDataJSONIndexByKey(datas []DataJSON, key string) (int, bool) {
	for i, d := range datas {
		if d.Key == key {
			return i, true
		}
	}
	return -1, false
}

// path ไปยัง fontoverride.json
func ReadfontsJSON(path string) ([]FontOverrideJSON, error) {
	d, err := os.ReadFile(path)
	if err != nil {
		return nil, ErrFileNotFound
	}
	var fontoverrides []FontOverrideJSON
	err = json.Unmarshal(d, &fontoverrides)
	if err != nil {
		return nil, errord.Errorf("error unmarshalling font overrides: %w", err)
	}
	return fontoverrides, nil
}

func ReadfontsJSONFromEmbedFS(embedFS embed.FS, embedPath string) ([]FontOverrideJSON, error) {
	d, err := embedFS.ReadFile(embedPath)
	if err != nil {
		return nil, ErrFileNotFound
	}
	var fontoverrides []FontOverrideJSON
	err = json.Unmarshal(d, &fontoverrides)
	if err != nil {
		return nil, errord.Errorf("error unmarshalling font overrides from embed: %w", err)
	}
	return fontoverrides, nil
}

func ReadTmplDir(path string) (Tmpl, error) {

	folderName := filepath.Base(path)
	tmplJsonPath := filepath.Join(path, "tmpl.json")
	tmplPdfPath := filepath.Join(path, "tmpl.pdf")

	tmplJson, err := parseTmplJSON(tmplJsonPath)
	if err != nil {
		return Tmpl{}, errord.Errorf("error reading tmpl.json: %w", err)
	}
	return Tmpl{
		tmplFolderPath: path,
		folderName:     folderName,
		tmplJSON:       &tmplJson,
		tmplPDFPath:    tmplPdfPath,
	}, nil
}

func ReadTmplDirFromEmbedFS(embedFS embed.FS, embedPath string) (Tmpl, error) {
	folderName := filepath.Base(embedPath)
	tmplJsonPath := filepath.Join(embedPath, "tmpl.json")
	tmplPdfPath := filepath.Join(embedPath, "tmpl.pdf")

	tmplJson, err := parseTmplJSONFromEmbedFS(embedFS, tmplJsonPath)
	if err != nil {
		return Tmpl{}, errord.Errorf("error reading tmpl.json from embed: %w", err)
	}
	return Tmpl{
		tmplFolderPath: embedPath,
		folderName:     folderName,
		tmplJSON:       &tmplJson,
		tmplPDFPath:    tmplPdfPath,
		embedFS:        &embedFS,
		embedPath:      embedPath,
	}, nil
}

func RemoveSpecialRuneInDataJSONSlice(src []DataJSON) []DataJSON {
	dest := make([]DataJSON, len(src))
	for i, s := range src {
		dest[i].Type = s.Type
		dest[i].Key = s.Key
		dest[i].WrapTextType = s.WrapTextType
		if s.Type == TypeText {
			dest[i].Val = RemoveSpecialRune(s.Val)
		} else {
			dest[i].Val = s.Val // TypeIMG ไม่ต้องลบ
		}
	}
	return dest
}

func RemoveSpecialRune(txt string) string {
	const carriageReturn = '\u000D'
	const lineFeed = '\u000A'
	var buff bytes.Buffer
	for _, r := range txt {
		if r == carriageReturn || r == lineFeed {
			continue
		}
		buff.WriteRune(r)
	}
	return buff.String()
}

// --- private ---

func parseTmplJSON(path string) (TmplJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TmplJSON{}, err
	}
	var objs TmplJSON
	err = json.Unmarshal(data, &objs)
	if err != nil {
		return TmplJSON{}, err
	}
	return objs, nil
}

func parseTmplJSONFromEmbedFS(embedFS embed.FS, path string) (TmplJSON, error) {
	data, err := embedFS.ReadFile(path)
	if err != nil {
		return TmplJSON{}, err
	}
	var objs TmplJSON
	err = json.Unmarshal(data, &objs)
	if err != nil {
		return TmplJSON{}, err
	}
	return objs, nil
}

func isInRune(r rune, vals []string) bool {
	if vals == nil || len(vals) <= 0 {
		return true
	}
	str := string(r)
	for _, val := range vals {
		if str == val {
			return true
		}
	}
	return false
}
