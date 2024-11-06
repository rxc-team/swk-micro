package charsetx

import (
	"bytes"
	"errors"
	"fmt"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

// Converts the specified charset to UTF-8.
// "chinese":             gbk,
// "csgb2312":            gbk,
// "csiso58gb231280":     gbk,
// "gb2312":              gbk,
// "gb_2312":             gbk,
// "gb_2312-80":          gbk,
// "gbk":                 gbk,
// "iso-ir-58":           gbk,
// "x-gbk":               gbk,
// "gb18030":             gb18030,
// "big5":                big5,
// "big5-hkscs":          big5,
// "cn-big5":             big5,
// "csbig5":              big5,
// "x-x-big5":            big5,
// "cseucpkdfmtjapanese": eucjp,
// "euc-jp":              eucjp,
// "x-euc-jp":            eucjp,
// "csiso2022jp":         iso2022jp,
// "iso-2022-jp":         iso2022jp,
// "csshiftjis":          shiftJIS,
// "ms932":               shiftJIS,
// "ms_kanji":            shiftJIS,
// "shift-jis":           shiftJIS,
// "shift_jis":           shiftJIS,
// "sjis":                shiftJIS,
// "windows-31j":         shiftJIS,
// "x-sjis":              shiftJIS,
func Decode(src []byte, charSet string) (string, error) {
	e, _ := charset.Lookup(charSet)
	if e == nil {
		return string(src), fmt.Errorf("invalid charset [%s]", charSet)
	}
	decodeStr, _, err := transform.Bytes(
		e.NewDecoder(),
		src,
	)
	if err != nil {
		return string(src), err
	}
	return string(decodeStr), nil
}

var encodings = []string{
	"sjis",
	"gbk",
	"utf-8",
}

// Converts to UTF-8.
// Charset (UTF-8, Shift-JIS, GBK) is automatically detected.
func DecodeAutoDetect(src []byte) (string, error) {
	for _, enc := range encodings {
		e, _ := charset.Lookup(enc)
		if e == nil {
			continue
		}
		var buf bytes.Buffer
		r := transform.NewWriter(&buf, e.NewDecoder())
		_, err := r.Write(src)
		if err != nil {
			continue
		}
		err = r.Close()
		if err != nil {
			continue
		}
		f := buf.Bytes()
		if isInvalidRune(f) {
			continue
		}
		if utf8.Valid(f) {
			if hasBom(f) {
				f = stripBom(f)
			}
			return string(f), nil
		}
	}
	return string(src), errors.New("could not determine character code")
}

var utf8bom = []byte{239, 187, 191}

// check have UTF-8 BOM
func hasBom(in []byte) bool {
	return bytes.HasPrefix(in, utf8bom)
}

// strip UTF-8 BOM
func stripBom(in []byte) []byte {
	return bytes.TrimPrefix(in, utf8bom)
}

func isInvalidRune(in []byte) bool {
	cb := in
	for len(cb) > 0 {
		if utf8.RuneStart(cb[0]) {
			r, size := utf8.DecodeRune(cb)
			if r == utf8.RuneError {
				return true
			}
			cb = cb[size:]
		} else {
			cb = cb[1:]
		}
	}
	return false
}
