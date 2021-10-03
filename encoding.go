package subtitles

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/saintfish/chardet"
)

// readAsUTF8 tries to convert io.Reader to UTF8
func readAsUTF8(r io.Reader) (string, error) {

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	if err != nil {
		return "", nil
	}

	return ConvertToUTF8(buf.Bytes())
}

func MustConvertToUTF8(b []byte) string {
	s, err := ConvertToUTF8(b)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

// ConvertToUTF8 returns a utf8 string
func ConvertToUTF8(b []byte) (string, error) {
	result, err := chardet.NewTextDetector().DetectBest(b)
	if err != nil {
		return "", fmt.Errorf("failed to detect character type: %v", err)
	}

	s := ""
	if result.Confidence > 50 {
		switch result.Charset {
		case "ISO-8859-1", "windows-1252":
			s = latin1toUTF8(b)
		case "UTF-16BE":
			s, _ = utf16ToUTF8(b[2:], true)
		case "UTF-16LE":
			s, _ = utf16ToUTF8(b[2:], false)
		case "UTF-8":
			if hasUTF8Marker(b) {
				s = string(b[3:])
			} else if utf8.ValidString(string(b)) {
				s = string(b)
			}
		default:
			return "", fmt.Errorf("unhandled chardet charset %q", result.Charset)
		}
	}

	// If we have little confidence in the result from chardet, we resort to our own checking.
	if s == "" {
		if hasUTF16BeMarker(b) {
			s, _ = utf16ToUTF8(b[2:], true)
		} else if hasUTF16LeMarker(b) {
			s, _ = utf16ToUTF8(b[2:], false)
		} else if hasUTF8Marker(b) {
			s = string(b[3:])
		} else if utf8.ValidString(string(b)) {
			s = string(b)
		} else if looksLikeLatin1(b) {
			s = latin1toUTF8(b)
		} else {
			s = string(b)
		}
	}

	str := normalizeLineFeeds(s)
	return str, nil
}

// NormalizeLineFeeds will return a string with \n as linefeeds
func normalizeLineFeeds(s string) string {
	if len(s) < 80 {
		return s
	}

	r := 0
	n := 0

	for i := 0; i < 80; i++ {
		if s[i] == '\r' {
			r++
		} else if s[i] == '\n' {
			n++
		}
	}

	if n == 0 && r > 0 {
		// older Mac files has \r linebreak
		return strings.Replace(s, "\r", "\n", -1)
	}

	return strings.Replace(s, "\r\n", "\n", -1)
}

func looksLikeLatin1(b []byte) bool {
	swe := float64(0)

	for i := 0; i < len(b); i++ {
		switch b[i] {
		case 0xe5, // å
			0xe4, // ä
			0xf6, // ö
			0xc4, // Ä
			0xc5, // Å
			0xd6: // Ö
			swe++
		}
	}

	// calc percent of swe letters
	pct := (swe / float64(len(b))) * 100
	if pct >= 1 {
		return true
	}

	if pct > 0 {
		//fmt.Printf("XXX %v %% swe letters, %v\n", pct, swe)
	}

	return false
}

func latin1toUTF8(in []byte) string {

	res := make([]rune, len(in))
	for i, b := range in {
		res[i] = rune(b)
	}
	return string(res)
}

func hasUTF8Marker(b []byte) bool {
	if len(b) < 3 {
		return false
	}
	if b[0] == 0xef && b[1] == 0xbb && b[2] == 0xbf {
		return true
	}
	return false
}
func hasUTF16BeMarker(b []byte) bool {
	if len(b) < 2 {
		return false
	}
	if b[0] == 0xfe && b[1] == 0xff {
		return true
	}
	return false
}

func hasUTF16LeMarker(b []byte) bool {
	if len(b) < 2 {
		return false
	}
	if b[0] == 0xff && b[1] == 0xfe {
		return true
	}
	return false
}

func utf16ToUTF8(b []byte, bigEndian bool) (string, error) {

	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		if bigEndian {
			u16s[0] = uint16(b[i+1]) + (uint16(b[i]) << 8)
		} else {
			u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		}
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}
