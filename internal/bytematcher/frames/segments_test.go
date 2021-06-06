package frames_test

import (
	"testing"

	. "github.com/ross-spencer/siegfried/internal/bytematcher/frames"
	. "github.com/ross-spencer/siegfried/internal/bytematcher/frames/tests"
)

const cost = 250000
const repetition = 4

// [BOF 0:test], [P 10-20:TESTY|YNESS], [S *:test|testy], [S 0:testy], [E 10-20:test|testy]
func TestSignatureOne(t *testing.T) {
	s := TestSignatures[0].Segment(8192, 2059, cost, repetition)
	if len(s) != 3 {
		t.Errorf("Segment fail: expecting 3 segments, got %d", len(s))
	}
	// [BOF 0:test], [P 10-20:TESTY|YNESS]
	if len(s[0]) != 2 {
		t.Errorf("Segment fail: expecting the first segment to have two frames, got %d", len(s[0]))
	}
	if s[0].Characterise() != BOFZero {
		t.Errorf("Characterise fail: expecting the first segment to be BOFZero, it is %v", s[0].Characterise())
	}
	pos := Position{4, 0, 1}
	if BOFLength(s[0], 64) != pos {
		t.Errorf("bofLength fail: expecting position %v, to equal %v", BOFLength(s[0], 64), pos)
	}
	// [S *:test|testy]
	if len(s[1]) != 1 {
		t.Errorf("Segment fail: expecting the second segment to have a single frame, got %d", len(s[0]))
	}
	if s[1].Characterise() != Succ {
		t.Errorf("Characterise fail: expecting the second segment to be Succ, it is %v", s[1].Characterise())
	}
	// the length in varLength reports the minimum, not the maximum length
	pos = Position{4, 0, 1}
	if VarLength(s[1], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", VarLength(s[1], 64), pos)
	}
	// [S 0:testy], [E 10-20:test|testy]
	if len(s[2]) != 2 {
		t.Errorf("Segment fail: expecting the last segment to have two frames, got %d", len(s[2]))
	}
	if s[2].Characterise() != EOFWindow {
		t.Errorf("Characterise fail: expecting the last segment to be eofWindow, it is %v", s[2].Characterise())
	}
	pos = Position{9, 0, 2}
	if VarLength(s[2], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", VarLength(s[2], 64), pos)
	}
}

// [BOF 0:test], [P 10-20:TESTY|YNESS], [P 0-1:TEST], [S 0:testy], [S *:test|testy], [E 0:23]
func TestSignatureTwo(t *testing.T) {
	s := TestSignatures[1].Segment(8192, 2059, cost, repetition)
	if len(s) != 3 {
		t.Errorf("Segment fail: expecting 3 segments, got %d", len(s))
	}
	// [BOF 0:test], [P 10-20:TESTY|YNESS], [P 0-1:TEST]
	if len(s[0]) != 3 {
		t.Errorf("Segment fail: expecting the first segment to have three frames, got %d", len(s[0]))
	}
	if s[0].Characterise() != BOFZero {
		t.Errorf("Characterise fail: expecting the first segment to be bofzero, it is %v", s[0].Characterise())
	}
	pos := Position{4, 0, 1}
	if BOFLength(s[0], 64) != pos {
		t.Errorf("bofLength fail: expecting position %v, to equal %v", BOFLength(s[0], 64), pos)
	}
	// [S 0:testy], [S *:test|testy]
	if len(s[1]) != 2 {
		t.Errorf("Segment fail: expecting the second segment to have two frames, got %d", len(s[1]))
	}
	if s[1].Characterise() != Succ {
		t.Errorf("Characterise fail: expecting the second segment to be succ, it is %v", s[1].Characterise())
	}
	pos = Position{9, 0, 2}
	if VarLength(s[1], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", BOFLength(s[1], 64), pos)
	}
}

// [BOF 0-5:a|b|c...|j], [P *:test]
func TestSignatureThree(t *testing.T) {
	s := TestSignatures[2].Segment(8192, 2059, cost, repetition)
	if len(s) != 2 {
		t.Errorf("Segment fail: expecting 2 segments, got %d", len(s))
	}
	// [BOF 0-5:a|b]
	if s[0].Characterise() != BOFWindow {
		t.Errorf("Characterise fail: expecting the first segment to be bofWindow, it is %v", s[0].Characterise())
	}
	pos := Position{1, 0, 1}
	if VarLength(s[0], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", VarLength(s[0], 64), pos)
	}
	// [P *:test]
	if len(s[1]) != 1 {
		t.Errorf("Segment fail: expecting the second segment to have one frame, got %d", len(s[1]))
	}
	if s[1].Characterise() != Prev {
		t.Errorf("Characterise fail: expecting the second segment to be prev, it is %v", s[1].Characterise())
	}
	pos = Position{4, 0, 1}
	if VarLength(s[1], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", VarLength(s[1], 64), pos)
	}
}

// [BOF 0:test], [P 10-20:TESTY|YNESS], [BOF *:test]
func TestSignatureFour(t *testing.T) {
	s := TestSignatures[3].Segment(8192, 2059, cost, repetition)
	if len(s) != 2 {
		t.Errorf("Segment fail: expecting 2 segments, got %d", len(s))
	}
	// [BOF 0:test], [P 10-20:TESTY|YNESS]
	if s[0].Characterise() != BOFZero {
		t.Errorf("Characterise fail: expecting the first segment to be bofWindow, it is %v", s[0].Characterise())
	}
	pos := Position{4, 0, 1}
	if BOFLength(s[0], 64) != pos {
		t.Errorf("bofLength fail: expecting position %v, to equal %v", BOFLength(s[0], 64), pos)
	}
	// [BOF *:test]
	if len(s[1]) != 1 {
		t.Errorf("Segment fail: expecting the second segment to have one frame, got %d", len(s[1]))
	}
	if s[1].Characterise() != BOFWild {
		t.Errorf("Characterise fail: expecting the second segment to be prev, it is %v", s[1].Characterise())
	}
	pos = Position{4, 0, 1}
	if VarLength(s[1], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", VarLength(s[1], 64), pos)
	}
}

func TestFmt418(t *testing.T) {
	s := TestFmts[418].Segment(2000, 500, cost, repetition)
	if len(s) != 2 {
		t.Errorf("fmt418 fail: expecting 2 segments, got %d", len(s))
	}
	if s[0].Characterise() != BOFZero {
		t.Errorf("fmt418 fail: expecting the first segment to be bofzero, got %v", s[0].Characterise())
	}
	pos := Position{14, 0, 1}
	if BOFLength(s[0], 2) != pos {
		t.Errorf("fmt418 fail: expecting the first segment to have pos %v, got %v", pos, BOFLength(s[0], 2))
	}
	if s[1].Characterise() != Prev {
		t.Errorf("fmt418 fail: expecting the second segment to be prev, got %v", s[1].Characterise())
	}
	pos = Position{33, 0, 2}
	if VarLength(s[1], 2) != pos {
		t.Errorf("fmt418 fail: expecting the second segment to have pos %v, got %v", pos, BOFLength(s[1], 2))
		t.Error(s[1])
	}
}
