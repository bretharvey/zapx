//  Copyright (c) 2026 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zap

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
)

func TestPostingsIteratorLocationLength(t *testing.T) {
	// A decoded frequency this large used to size the location slices and
	// abort the process. The location payload is what bounds the allocation.
	freq := uint64(1) << 42
	freqNorm := binary.AppendUvarint(nil, encodeFreqHasLocs(freq, true))
	freqNorm = binary.AppendUvarint(freqNorm, 0)

	numLocsBytes := uint64(1) << 40
	loc := binary.AppendUvarint(nil, numLocsBytes)

	// Offset 0 means the section is absent.
	mem := []byte{0}
	freqOffset := uint64(len(mem))
	mem = append(mem, testPostingChunk(freqNorm)...)
	locOffset := uint64(len(mem))
	mem = append(mem, testPostingChunk(loc)...)

	postings := roaring.NewBitmap()
	postings.Add(0)

	pl := &PostingsList{
		sb: &SegmentBase{
			mem:        mem,
			fileReader: &FileReader{},
		},
		postings:   postings,
		chunkSize:  1024,
		freqOffset: freqOffset,
		locOffset:  locOffset,
	}

	itr := pl.iterator(true, true, true, nil)
	got, err := itr.Next()
	if err == nil {
		t.Fatalf("expected location length error, got posting freq %d", got.Frequency())
	}
	if !strings.Contains(err.Error(), "exceeds remaining chunk bytes") {
		t.Fatalf("expected location length error, got: %v", err)
	}
}

func testPostingChunk(payload []byte) []byte {
	buf := binary.AppendUvarint(nil, 1)
	buf = binary.AppendUvarint(buf, uint64(len(payload)))
	return append(buf, payload...)
}
