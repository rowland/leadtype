// Copyright 2013 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package wordbreaking

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-text/typesetting/segmenter"
)

func markedRuneAttributes(text string) []Flags {
	flags := make([]Flags, len(text))
	MarkRuneAttributes(text, flags)
	return flags
}

func offsetsWithFlag(flags []Flags, flag Flags) []int {
	offsets := make([]int, 0)
	for offset, attrs := range flags {
		if attrs&flag != 0 {
			offsets = append(offsets, offset)
		}
	}
	return offsets
}

func testRuneByteOffsets(text string) []int {
	offsets := make([]int, 0, utf8.RuneCountInString(text)+1)
	for offset := range text {
		offsets = append(offsets, offset)
	}
	return append(offsets, len(text))
}

func testUAX14LineBreakOffsets(text string) []int {
	byteOffsets := testRuneByteOffsets(text)
	var seg segmenter.Segmenter
	seg.InitWithString(text)
	iter := seg.LineIterator()
	offsets := make([]int, 0)
	for iter.Next() {
		line := iter.Line()
		runeOffset := line.Offset + len(line.Text)
		if runeOffset < len(byteOffsets)-1 {
			offsets = append(offsets, byteOffsets[runeOffset])
		}
	}
	return offsets
}

func textWithBreakMarkers(text string, offsets []int) string {
	breaks := make(map[int]struct{}, len(offsets))
	for _, offset := range offsets {
		breaks[offset] = struct{}{}
	}
	var marked strings.Builder
	marked.Grow(len(text) + len(offsets))
	for offset, r := range text {
		if _, ok := breaks[offset]; ok {
			marked.WriteByte('|')
		}
		marked.WriteRune(r)
	}
	return marked.String()
}

func TestMarkRuneAttributes(t *testing.T) {
	const quick = "The quick red fox jumps over the lazy brown dog."
	var quickFlags = []Flags{
		CharStop | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // The
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // quick
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // red
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // fox
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // jumps
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // over
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // the
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // lazy
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // brown
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, // dog.
	}
	var flags [len(quick)]Flags

	MarkRuneAttributes(quick, flags[:])
	for i := range flags {
		if flags[i] != quickFlags[i] {
			t.Errorf("Expecting %d, got %d at index %d", quickFlags[i], flags[i], i)
		}
	}
}

func TestMarkRuneAttributes_with_hyphens(t *testing.T) {
	const hyphenTest = "Word-breaking test with regular and soft\u00ADhyphens."
	var hyphenFlags = []Flags{
		CharStop | WordStop, CharStop, CharStop, CharStop, CharStop, // Word-
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // breaking
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // test
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // with
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // regular
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // and
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, 0, // soft-
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, // hyphens.
	}
	var flags [len(hyphenTest)]Flags

	MarkRuneAttributes(hyphenTest, flags[:])
	for i := range flags {
		if flags[i] != hyphenFlags[i] {
			t.Errorf("Expecting %d, got %d at index %d", hyphenFlags[i], flags[i], i)
		}
	}
}

func TestMarkRuneAttributes_cjkBreaks(t *testing.T) {
	const text = "中文かなABC"
	flags := make([]Flags, len(text))

	MarkRuneAttributes(text, flags)

	if flags[0] != CharStop|WordStop {
		t.Fatalf("first rune flags = %08b, want %08b", flags[0], CharStop|WordStop)
	}
	for _, idx := range []int{3, 6, 9} {
		if flags[idx]&SoftBreak == 0 {
			t.Fatalf("expected soft break at byte offset %d, got %08b", idx, flags[idx])
		}
		if flags[idx]&WordStop == 0 {
			t.Fatalf("expected word stop at byte offset %d, got %08b", idx, flags[idx])
		}
	}
}

func TestMarkRuneAttributes_UTF8ByteOffsets(t *testing.T) {
	const text = "Aé中😀B"
	flags := markedRuneAttributes(text)

	if got, want := offsetsWithFlag(flags, CharStop), []int{0, 1, 3, 6, 10}; !slices.Equal(got, want) {
		t.Fatalf("CharStop byte offsets = %v, want %v", got, want)
	}
	if got, want := offsetsWithFlag(flags, SoftBreak), []int{3, 6, 10}; !slices.Equal(got, want) {
		t.Fatalf("SoftBreak byte offsets = %v, want %v", got, want)
	}
	for _, offset := range []int{2, 4, 5, 7, 8, 9} {
		if flags[offset] != 0 {
			t.Errorf("UTF-8 continuation byte at offset %d has flags %08b, want zero", offset, flags[offset])
		}
	}
}

func TestMarkRuneAttributes_BreakControls(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		softBreaks []int
	}{
		{name: "space", text: "word break", softBreaks: []int{5}},
		{name: "hyphen-minus", text: "well-being", softBreaks: []int{5}},
		{name: "soft hyphen", text: "well\u00adbeing", softBreaks: []int{6}},
		{name: "non-breaking hyphen", text: "well\u2011being"},
		{name: "no-break space", text: "well\u00a0being"},
		{name: "zero-width space", text: "well\u200bbeing", softBreaks: []int{7}},
		{name: "mixed Latin and CJK", text: "ABC中文DEF", softBreaks: []int{3, 6, 9}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := markedRuneAttributes(tt.text)
			if got := offsetsWithFlag(flags, SoftBreak); !slices.Equal(got, tt.softBreaks) {
				t.Fatalf("SoftBreak byte offsets = %v, want %v", got, tt.softBreaks)
			}
		})
	}
}

func TestMarkRuneAttributes_ThaiDictionaryBoundary(t *testing.T) {
	const text = "เขาไป"
	flags := markedRuneAttributes(text)
	breakOffset := strings.Index(text, "ไป")
	if flags[breakOffset]&SoftBreak == 0 || flags[breakOffset]&WordStop == 0 {
		t.Fatalf("Thai dictionary boundary at byte offset %d has flags %08b", breakOffset, flags[breakOffset])
	}
}

func TestMarkRuneAttributes_AugmentsCallerFlags(t *testing.T) {
	flags := make([]Flags, len("ab"))
	flags[0] = NoBreak
	flags[1] = Invalid

	MarkRuneAttributes("ab", flags)

	if flags[0]&NoBreak == 0 || flags[0]&CharStop == 0 || flags[0]&WordStop == 0 {
		t.Fatalf("first byte flags = %08b, want caller NoBreak plus rune attributes", flags[0])
	}
	if flags[1]&Invalid == 0 || flags[1]&CharStop == 0 {
		t.Fatalf("second byte flags = %08b, want caller Invalid plus CharStop", flags[1])
	}
}

func TestMarkRuneAttributes_KoreanUsesUAX14Boundaries(t *testing.T) {
	const text = "한국어 예문 이른 오후가 되면"
	want := testUAX14LineBreakOffsets(text)
	if got := offsetsWithFlag(markedRuneAttributes(text), SoftBreak); !slices.Equal(got, want) {
		t.Fatalf("Korean SoftBreak byte offsets = %v, want UAX #14 boundaries %v", got, want)
	}
}

func TestMarkRuneAttributes_KoreanBoundaryBaseline(t *testing.T) {
	const text = "한국어 예문: 이른 오후가 되면 골목의 카페에는 사람들이 늘어나고, 창가에 앉습니다."
	leadtypeOffsets := offsetsWithFlag(markedRuneAttributes(text), SoftBreak)
	uaxOffsets := testUAX14LineBreakOffsets(text)
	leadtypeMarked := textWithBreakMarkers(text, leadtypeOffsets)
	uaxMarked := textWithBreakMarkers(text, uaxOffsets)

	const wantUAX14 = "한|국|어 |예|문: |이|른 |오|후|가 |되|면 |골|목|의 |카|페|에|는 |사|람|들|이 |늘|어|나|고, |창|가|에 |앉|습|니|다."
	if leadtypeMarked != wantUAX14 {
		t.Errorf("Leadtype Korean boundaries = %q, want %q", leadtypeMarked, wantUAX14)
	}
	if uaxMarked != wantUAX14 {
		t.Errorf("UAX #14 Korean boundaries = %q; update pinned baseline", uaxMarked)
	}
	t.Logf("Leadtype: %s", leadtypeMarked)
	t.Logf("UAX #14: %s", uaxMarked)
}

func TestMarkRuneAttributes_MatchesUAX14Boundaries(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "Latin", text: "Leadtype wraps words and punctuation."},
		{name: "kana", text: "かなカナ文章"},
		{name: "Bopomofo", text: "中文ㄅㄆㄇ中文"},
		{name: "numeric", text: "価格は123.45%です。"},
		{name: "emoji ZWJ", text: "中👩‍👩‍👧‍👦文"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := offsetsWithFlag(markedRuneAttributes(tt.text), SoftBreak)
			want := testUAX14LineBreakOffsets(tt.text)
			if !slices.Equal(got, want) {
				t.Fatalf("SoftBreak byte offsets = %v, want UAX #14 boundaries %v", got, want)
			}
		})
	}
}

func TestMarkRuneAttributes_ProhibitsBreakBeforeClosingPunctuationAndNonstarters(t *testing.T) {
	for _, r := range "，、。！？）】」』》〉々" {
		t.Run(string(r), func(t *testing.T) {
			text := "中" + string(r) + "文"
			offset := len("中")
			flags := markedRuneAttributes(text)
			if flags[offset]&SoftBreak != 0 {
				t.Fatalf("SoftBreak before %q at byte offset %d must be prohibited", r, offset)
			}
		})
	}
}

func TestMarkRuneAttributes_ProhibitsBreakAfterOpeningPunctuation(t *testing.T) {
	for _, r := range "（【「『《〈" {
		t.Run(string(r), func(t *testing.T) {
			text := string(r) + "中文"
			offset := len(string(r))
			flags := markedRuneAttributes(text)
			if flags[offset]&SoftBreak != 0 {
				t.Fatalf("SoftBreak after %q at byte offset %d must be prohibited", r, offset)
			}
		})
	}
}

func TestMarkRuneAttributes_ProhibitsBreakBeforeCombiningMark(t *testing.T) {
	const text = "中\u0301文"
	offset := len("中")
	flags := markedRuneAttributes(text)
	if flags[offset]&SoftBreak != 0 {
		t.Fatalf("SoftBreak before combining mark at byte offset %d must be prohibited", offset)
	}
}

func TestMarkRuneAttributes_HonorsWordJoiner(t *testing.T) {
	const text = "中\u2060文"
	flags := markedRuneAttributes(text)
	for _, offset := range []int{len("中"), len("中\u2060")} {
		if flags[offset]&SoftBreak != 0 {
			t.Errorf("SoftBreak adjacent to WORD JOINER at byte offset %d must be prohibited", offset)
		}
	}
}

func TestMarkRuneAttributes_HonorsNarrowNoBreakSpace(t *testing.T) {
	const text = "中\u202f文"
	flags := markedRuneAttributes(text)
	for _, offset := range []int{len("中"), len("中\u202f")} {
		if flags[offset]&SoftBreak != 0 {
			t.Errorf("SoftBreak adjacent to NARROW NO-BREAK SPACE at byte offset %d must be prohibited", offset)
		}
	}
}

func TestMarkRuneAttributes_MandatoryBreaks(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		offset int
	}{
		{name: "LF", text: "alpha\nbeta", offset: len("alpha\n")},
		{name: "CRLF", text: "alpha\r\nbeta", offset: len("alpha\r\n")},
		{name: "NEL", text: "alpha\u0085beta", offset: len("alpha\u0085")},
		{name: "line separator", text: "alpha\u2028beta", offset: len("alpha\u2028")},
		{name: "paragraph separator", text: "alpha\u2029beta", offset: len("alpha\u2029")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := markedRuneAttributes(tt.text)
			if got := offsetsWithFlag(flags, MandatoryBreak); !slices.Equal(got, []int{tt.offset}) {
				t.Fatalf("MandatoryBreak byte offsets = %v, want [%d]", got, tt.offset)
			}
			if flags[tt.offset]&SoftBreak == 0 {
				t.Fatalf("mandatory boundary at byte offset %d is missing SoftBreak", tt.offset)
			}
		})
	}
}

func TestMarkRuneAttributes_TrailingMandatoryBreakHasNoEOFSlot(t *testing.T) {
	const text = "alpha\n"
	if got := offsetsWithFlag(markedRuneAttributes(text), MandatoryBreak); len(got) != 0 {
		t.Fatalf("MandatoryBreak byte offsets = %v, want none before EOF", got)
	}
}

func BenchmarkMarkRuneAttributes(b *testing.B) {
	corpora := map[string]string{
		"Latin":    "The quick red fox jumps over the lazy brown dog.",
		"Chinese":  "汉语示例：清晨的街道刚刚苏醒，卖早点的小店升起热气，行人提着公文包快步穿过路口。",
		"Japanese": "日本語の例文：朝の駅前では、通勤する人々が改札へ向かい、パン屋からは焼きたての香りが流れてきます。",
		"Korean":   "한국어 예문: 이른 오후가 되면 골목의 카페에는 사람들의 대화가 조금씩 늘어납니다.",
		"Thai":     "ตัวอย่างภาษาไทย: ในช่วงบ่ายอากาศเริ่มอ่อนลง ร้านกาแฟเล็ก ๆ ริมถนนมีผู้คนแวะมานั่งพัก",
		"Mixed":    "中\u2060文 A\u00a0B A\u202fB 👩‍👩‍👧‍👦 ภาษาไทย",
	}

	for name, text := range corpora {
		b.Run(name, func(b *testing.B) {
			flags := make([]Flags, len(text))
			MarkRuneAttributes(text, flags) // Warm lazy Thai dictionary initialization.
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				MarkRuneAttributes(text, flags)
			}
		})
	}
}
