package libanreport

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/oneplus1000/errord"
)

func TestGenPdf(t *testing.T) {
	err := testGenPdf()
	if err != nil {
		t.Errorf("testGenPdf failed: %v", err)
	}
}

func testGenPdf() error {

	//สร้างโฟลเดอร์สำหรับเก็บผลลัพธ์ ถ้ายังไม่มี
	if _, err := os.Stat("testing_output"); os.IsNotExist(err) {
		err = os.Mkdir("testing_output", 0755)
		if err != nil {
			return errord.Errorf("error creating output directory: %w", err)
		}
	}

	//code จริงเริ่มที่นี่
	outputPath := filepath.Join("testing_output", "test06.pdf")
	tmplPath := filepath.Join("testing", "test06.tmpl")
	tmpl, err := ReadTmplDir(tmplPath)
	if err != nil {
		return errord.Errorf("error reading template directory: %w", err)
	}
	fontOverrides, err := ReadfontsJSON(filepath.Join("testing", "fontoverride.json"))
	if err != nil {
		return errord.Errorf("error reading fontoverride.json: %w", err)
	}

	data := RemoveSpecialRuneInDataJSONSlice([]DataJSON{
		{Type: 1, Key: "cusLicence", Val: "ผค 555 กท", WrapTextType: WrapTextTypeNewLine},
		{Type: 1, Key: "cusAddress", Val: "1259/67 หมู่บ้านเสนากรีนวิลล์ รามอินทรา ถนนพระยาสุเรนทร์ แขวงบางชัน เขตคลองสามวา กรุงเทพฯ 10510 หมู่บ้านเสนากรีนวิลล์ รามอินทรา ถนนพระยาสุเรนทร์ แขวงบางชัน เขตคลองสามวา กรุงเทพฯ 10510", WrapTextType: WrapTextTypeNewLine},
		//{Type: 1, Key: "c1Company", Val: "บริษัท เมืองไทยประกันภัยป่า จำกัด (มหาชน)"},
		//{Type: 1, Key: "c1Garage", Val: "ป. 1 ซ่อมห้าง"},
		//{Type: 1, Key: "cusModel", Val: "FORTUNER 2022"},
		//{Type: 1, Key: "cusLicence", Val: "3 ขฎ 597 กทม"},
		//{Type: 1, Key: "cusChassis", Val: "MR0AB3GS702576035"},
		//{Type: 1, Key: "cusYear", Val: "2022"},
		//{Type: 1, Key: "cusExpire", Val: "30 มิถุนายน 2568"},
		//{Type: 1, Key: "cusChassis", Val: "MR0AB3GS702576035"},
	})

	f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errord.Errorf("error opening output file: %w", err)
	}

	err = GenPdf(
		tmpl,
		data,
		fontOverrides,
		f,
	)
	if err != nil {
		return fmt.Errorf("error generating PDF: %w", err)
	}
	return nil
}
