package subtitles_test

import (
	"io/ioutil"
	"testing"

	"github.com/martinlindhe/subtitles"
)

func TestParse(t *testing.T) {
	wantSRT := []string{
		`1
00:55:40,920 --> 00:55:44,760
Jag vill ha tillbaka den här sen.

`,
		`2
00:55:46,800 --> 00:55:50,800
-God natt.
-Kom nu. - God natt.

`,
		`3
00:55:51,800 --> 00:55:55,800
Vi ses i morgon och tar nya tag.

`,
		`4
00:58:25,000 --> 00:58:29,280
Textning: Anders Kaage
Svensk Medietext för SVT

`}
	for _, file := range []string{"srt/sample.srt", "srt/sample.latin1.srt", "srt/sample.latin1.dos.srt"} {
		path := "./testdata/" + file
		b, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatalf("ioutil.ReadFile(%s): %v", path, err)
		}
		sub, err := subtitles.Parse(b)
		if err != nil {
			t.Fatalf("%s Parse(): %v", file, err)
		}
		if got, want := len(sub.Captions), 4; got != want {
			t.Fatalf("%s sub.Captions() got %d, want %d", file, got, want)
		}
		for i, capt := range sub.Captions {
			if got, want := capt.AsSRT(), wantSRT[i]; got != want {
				t.Errorf("%s Caption %d got [%s], want [%s]", file, i, got, want)
			}
		}
	}
}
