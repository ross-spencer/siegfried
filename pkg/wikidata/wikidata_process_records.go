// Copyright 2020 Ross Spencer, Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

// Helper functions for creating the sets of signatures that will be
// processed into the Wikidata identifier. As Wikidata entries are
// processed records are either created new, or appended/updated.

package wikidata

import (
	"github.com/richardlehane/siegfried/pkg/wikidata/internal/mappings"

	"github.com/ross-spencer/spargo/pkg/spargo"
)

// wikidataRecord provides an alias for the mappings.Wikidata object.
type wikidataRecord = mappings.Wikidata

// newRecord creates a Wikidata record with the values received from
// Wikidata itself.
func newRecord(wikidataItem map[string]spargo.Item, addSigs bool) wikidataRecord {
	wd := wikidataRecord{}
	wd.ID = getID(wikidataItem[uriField].Value)
	wd.Name = wikidataItem[formatLabelField].Value
	wd.URI = wikidataItem[uriField].Value
	wd.PRONOM = append(wd.PRONOM, wikidataItem[puidField].Value)
	if wikidataItem[extField].Value != "" {
		wd.Extension = append(wd.Extension, wikidataItem[extField].Value)
	}
	wd.Mimetype = append(wd.Mimetype, wikidataItem[mimeField].Value)
	if wikidataItem[signatureField].Value != "" {
		if !addSigs {
			// Pre-processing has determined that no particular
			// heuristic will help us here and so let's make sure we can
			// report on that at the end, as well as exit early.
			addLinting(wd.URI, heuWDE01)
			wd.DisableSignatures()
			return wd
		}
		sig := Signature{}
		sig.Provenance = parseProvenance(wikidataItem[referenceField].Value)
		sig.Date = wikidataItem[dateField].Value

		wd.Signatures = append(wd.Signatures, sig)
		bs := newByteSequence(wikidataItem)
		wd.Signatures[0].ByteSequences = append(
			wd.Signatures[0].ByteSequences, bs)
	}
	return wd
}

// updateRecord manages a format record's repeating properties.
// exceptions and adds them to the list if it doesn't already exist.
func updateRecord(wikidataItem map[string]spargo.Item, wd wikidataRecord) wikidataRecord {
	if contains(wd.PRONOM, wikidataItem[puidField].Value) == false {
		wd.PRONOM = append(wd.PRONOM, wikidataItem[puidField].Value)
	}
	if contains(wd.Extension, wikidataItem[extField].Value) == false &&
		wikidataItem[extField].Value != "" {
		wd.Extension = append(wd.Extension, wikidataItem[extField].Value)
	}
	if contains(wd.Mimetype, wikidataItem[mimeField].Value) == false {
		wd.Mimetype = append(wd.Mimetype, wikidataItem[mimeField].Value)
	}
	if wikidataItem[signatureField].Value != "" {
		if !wd.SignaturesDisabled() {
			lintingErr := updateSequences(wikidataItem, &wd)
			// WIKIDATA FUTURE: If we can re-organize the signatures in
			// Wikidata so that they are better encapsulated from each
			// other then we don't need to be as strict about not
			// processing the value. Right now, there's not enough
			// consistency in records that mix signatures with multiple
			// sequences, types, offsets and so forth.
			if lintingErr != nle {
				wd.Signatures = nil
				wd.DisableSignatures()
				addLinting(wd.URI, lintingErr)
			}
		}
	}
	return wd
}
