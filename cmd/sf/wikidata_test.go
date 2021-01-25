package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/wikidata"
	"github.com/richardlehane/siegfried/pkg/writer"
)

// Path components associated with the Roy command folder.
const wikidataTestDefinitions = "wikidata-test-definitions"
const wikidataDefinitionsBaseDir = "definitionsBaseDir"

var royTestData = filepath.Join("..", "roy", "data")

// Path components within the Siegfried command folder.
const wikidataNamespace = "wikidata"
const siegfriedTestData = "testdata"
const wikidataTestData = "wikidata"
const wikidataPRONOMSkeletons = "pro"
const wikidataCustomSkeletons = "wd"
const wikidataArcSkeletons = "arc"
const wikidataExtensionMismatches = "ext_mismatch"
const wikidataContainerMatches = "container"
const wikidataCuriosityMatches = "curiosities"

var (
	wikidataDefinitions = flag.String(
		wikidataDefinitionsBaseDir,
		royTestData,
		"Creates an flag var that is compatible with the config functions...",
	)
)

var wdSiegfried *siegfried.Siegfried

func setupWikidata(pronomx bool) error {
	wdSiegfried = siegfried.New()
	config.SetHome(*wikidataDefinitions)
	config.SetWikidataDefinitions(wikidataTestDefinitions)
	opts := []config.Option{config.SetWikidataNamespace()}
	if pronomx != true {
		opts = append(opts, config.SetWikidataNoPRONOM())
		opts = append(opts, config.SetWikidataSourceFieldOff())
	} else {
		opts = append(opts, config.SetWikidataPRONOM())
	}
	identifier, err := wikidata.New(opts...)
	if err != nil {
		return err
	}
	wdSiegfried.Add(identifier)
	return nil
}

// identificationTests provides our structure for table driven tests.
type identificationTests struct {
	fname          string
	qid            string
	extMatch       bool
	byteMatch      bool
	containerMatch bool
	error          bool
}

var skeletonSamples = []identificationTests{
	identificationTests{
		filepath.Join(wikidataPRONOMSkeletons, "fmt-11-signature-id-58.png"),
		"Q178051", true, true, false, false},
	identificationTests{
		filepath.Join(wikidataPRONOMSkeletons, "fmt-279-signature-id-295.flac"),
		"Q27881556", true, true, false, false},
	identificationTests{
		filepath.Join(wikidataCustomSkeletons, "Q10287816.gz"),
		"Q10287816", true, true, false, false},
	identificationTests{
		filepath.Join(wikidataCustomSkeletons, "Q28205479.info"),
		"Q28205479", true, true, false, false},
}

// Rudimentary consts that can help us determine the method of
// identification. Can also add "container name" here for when we want
// to validate PRONOM alongside Wikidata.
const extensionMatch = "extension match"
const byteMatch = "byte match"
const extensionMismatch = "extension mismatch"
const containerMatch = "container name"

// TestWikidataBasic will perform some rudimentary tests using some
// simple Skeleton files and the Wikidata identifier without PRONOM.
func TestWikidataBasic(t *testing.T) {
	err := setupWikidata(false)
	if err != nil {
		t.Error(err)
	}
	for _, test := range skeletonSamples {
		path := filepath.Join(siegfriedTestData, wikidataTestData, test.fname)
		siegfriedRunner(path, test, t)
	}
	wdSiegfried = nil
}

var archiveSamples = []identificationTests{
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "fmt-289-signature-id-305.warc"),
		"Q7978505", true, true, false, false},
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "fmt-410-signature-id-580.arc"),
		"Q27824065", true, true, false, false},
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "x-fmt-219-signature-id-525.arc"),
		"Q27824060", true, true, false, false},
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "x-fmt-265-signature-id-265.tar"),
		"Q283579", true, true, false, false},
	identificationTests{
		filepath.Join(wikidataArcSkeletons, "x-fmt-266-signature-id-201.gz"),
		"Q10287816", true, true, false, false},
}

func TestArchives(t *testing.T) {
	err := setupWikidata(true)
	if err != nil {
		t.Error(err)
	}
	for _, test := range archiveSamples {
		path := filepath.Join(siegfriedTestData, wikidataTestData, test.fname)
		siegfriedRunner(path, test, t)
	}
	wdSiegfried = nil
}

var extensionMismatchSamples = []identificationTests{
	identificationTests{
		filepath.Join(wikidataExtensionMismatches, "fmt-11-signature-id-58.jpg"),
		"Q178051", false, true, false, false},
	identificationTests{
		filepath.Join(wikidataExtensionMismatches, "fmt-279-signature-id-295.wav"),
		"Q27881556", false, true, false, false},
}

func TestExtensionMismatches(t *testing.T) {
	err := setupWikidata(false)
	if err != nil {
		t.Error(err)
	}
	for _, test := range extensionMismatchSamples {
		path := filepath.Join(siegfriedTestData, wikidataTestData, test.fname)
		siegfriedRunner(path, test, t)
	}
	wdSiegfried = nil
}

var containerSamples = []identificationTests{
	identificationTests{
		filepath.Join(wikidataContainerMatches, "fmt-292-container-signature-id-8010.odp"),
		"Q27203973", true, true, true, false},
	identificationTests{
		filepath.Join(wikidataContainerMatches, "fmt-482-container-signature-id-14000.ibooks"),
		"Q49988096", true, true, true, false},
	identificationTests{
		filepath.Join(wikidataContainerMatches, "fmt-680-container-signature-id-22120.ppp"),
		"Q47520869", true, true, true, false},
	identificationTests{
		filepath.Join(wikidataContainerMatches, "fmt-998-container-signature-id-32000.ora"),
		"Q747906", true, true, true, false},
}

func TestContainers(t *testing.T) {
	err := setupWikidata(true)
	if err != nil {
		t.Error(err)
	}
	for _, test := range containerSamples {
		path := filepath.Join(siegfriedTestData, wikidataTestData, test.fname)
		siegfriedRunner(path, test, t)
	}
	wdSiegfried = nil
}

var curiositySamples = []identificationTests{
	identificationTests{
		// curiosity.1 should match with Q000000 in the sample signature
		// file which has source information including lots of strange
		// characters as well as emoji.
		filepath.Join(wikidataCuriosityMatches, "curiosity.1"),
		"Q000000", true, true, false, false},
}

func TestCurious(t *testing.T) {
	err := setupWikidata(true)
	if err != nil {
		t.Error(err)
	}
	for _, test := range curiositySamples {
		path := filepath.Join(siegfriedTestData, wikidataTestData, test.fname)
		siegfriedRunner(path, test, t)
	}
	wdSiegfried = nil
}

func siegfriedRunner(path string, test identificationTests, t *testing.T) {
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("failed to open %v, got: %v", path, err)
	}
	defer file.Close()
	res, err := wdSiegfried.Identify(file, path, "")
	if err != nil && !test.error {
		t.Fatal(err)
	}
	if len(res) > 1 {
		t.Errorf("Match length greater than one: '%d'", len(res))
	}
	namespace := res[0].Values()[0]
	if namespace != wikidataNamespace {
		t.Errorf("Namespace error, expected: '%s' received: '%s'", wikidataNamespace, namespace)
	}
	id := res[0].Values()[1]
	permalink := res[0].Values()[4]
	basis := res[0].Values()[6]
	warning := res[0].Values()[7]
	if id != test.qid {
		t.Errorf(
			"QID match different than anticipated: '%s' expected '%s'",
			id,
			test.qid,
		)
	}
	const placeholderPermalink = "https://www.wikidata.org/w/index.php?oldid=1287431117&title=Q12345"
	if permalink != placeholderPermalink {
		t.Errorf("There has been a problem parsing the permalink for '%s' from Wikidata/Wikiprov: %s",
			test.qid,
			permalink,
		)
	}

	if test.extMatch && !strings.Contains(basis, extensionMatch) {
		t.Errorf("Extension match not returned by identifier: %s", basis)
	}
	if test.byteMatch && !strings.Contains(basis, byteMatch) {
		t.Errorf("Byte match not returned by identifier: %s", basis)
	}
	if test.containerMatch && !strings.Contains(basis, containerMatch) {
		t.Errorf("Container match not returned by identifier: %s", basis)
	}
	if !test.extMatch && !strings.Contains(warning, extensionMismatch) {
		t.Errorf("Expected an extension mismatch but it wasn't returned: %s", warning)
	}

	// Implement a basic Writer test for some of the data coming out of
	// the Wikidata identifier. CSV and YAML will need a little more
	// thought.
	var w writer.Writer
	buf := new(bytes.Buffer)
	w = writer.JSON(buf)
	w.Head("path/to/file", time.Now(), time.Now(), [3]int{0, 0, 0}, wdSiegfried.Identifiers(), wdSiegfried.Fields(), "md5")
	w.File("testName", 10, "testMod", []byte("d41d8c"), nil, res)
	w.Tail()
	if !json.Valid([]byte(buf.String())) {
		t.Errorf("Output from JSON writer is invalid: %s", buf.String())
	}
}
