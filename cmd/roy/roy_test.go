package main

import (
	"flag"
	"testing"

	"github.com/ross-spencer/siegfried"
	"github.com/ross-spencer/siegfried/pkg/config"
	"github.com/ross-spencer/siegfried/pkg/loc"
	"github.com/ross-spencer/siegfried/pkg/mimeinfo"
	"github.com/ross-spencer/siegfried/pkg/pronom"
	"github.com/ross-spencer/siegfried/pkg/sets"
	wd "github.com/ross-spencer/siegfried/pkg/wikidata"
)

var testhome = flag.String("home", "data", "override the default home directory")

func TestDefault(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New()
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoc(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	l, err := loc.New(config.SetLOC(""))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(l)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTika(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	m, err := mimeinfo.New(config.SetMIMEInfo("tika"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFreedesktop(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	m, err := mimeinfo.New(config.SetMIMEInfo("freedesktop"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWikidata(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	config.SetWikidataDefinitions("wikidata-test-definitions")
	m, err := wd.New(config.SetWikidataNamespace())
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPronomTikaLoc(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New(config.Clear())
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
	m, err := mimeinfo.New(config.SetMIMEInfo("tika"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
	l, err := loc.New(config.SetLOC(""))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(l)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeluxe(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New(config.Clear())
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
	m, err := mimeinfo.New(config.SetMIMEInfo("tika"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
	f, err := mimeinfo.New(config.SetMIMEInfo("freedesktop"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(f)
	if err != nil {
		t.Fatal(err)
	}
	l, err := loc.New(config.SetLOC(""))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(l)
	if err != nil {
		t.Fatal(err)
	}
}

func TestArchivematica(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New(
		config.SetName("archivematica"),
		config.SetExtend(sets.Expand("archivematica-fmt2.xml,archivematica-fmt3.xml,archivematica-fmt4.xml,archivematica-fmt5.xml")))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
}
