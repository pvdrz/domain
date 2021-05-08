package text

import (
	"bytes"
	"math"

	"github.com/pvdrz/domain/lib/doc"
)

const termSize = 3

type term [termSize]byte

func forEachTerm(slice []byte, f func(term)) {
	slice = bytes.ToLower(slice)
	pos := 0

	for pos+termSize <= len(slice) {
		var term term
		copy(term[:], slice[pos:pos+termSize])

		f(term)

		pos += 1
	}
}

type countNum uint32
type scoreNum float32

type countKey struct {
	term term
	id   doc.DocID
}

type Index struct {
	termCounts     map[countKey]countNum
	documentCounts map[term]countNum
	maxTermCounts  map[doc.DocID]countNum
}

func NewIndex() Index {
	return Index{
		termCounts:     make(map[countKey]countNum),
		documentCounts: make(map[term]countNum),
		maxTermCounts:  make(map[doc.DocID]countNum),
	}
}

func (index *Index) Insert(id doc.DocID, document *doc.Doc) {
	count := make(map[term]countNum)

	f := func(term term) {
		count[term] += 1
	}

	forEachTerm([]byte(document.Title), f)
	for _, author := range document.Authors {
		forEachTerm([]byte(author), f)
	}
	for _, keyword := range document.Keywords {
		forEachTerm([]byte(keyword), f)
	}

	maxCount := countNum(0)

	for term, count := range count {
		index.termCounts[countKey{term: term, id: id}] = count

		index.documentCounts[term] += 1

		if count > maxCount {
			maxCount = count
		}
	}

	index.maxTermCounts[id] = maxCount
}

func (index *Index) Search(query []byte) []doc.DocID {
	const Max = 5

	var scores [Max]scoreNum
	var matches [Max]doc.DocID
	if len(query) < termSize {
		return matches[0:0]
	}

	total := scoreNum(len(index.maxTermCounts))
	foundCount := 0

	for id, maxCount := range index.maxTermCounts {
		maxCount := scoreNum(maxCount)
		score := scoreNum(0.0)

		forEachTerm(query, func(term term) {
			termCount := index.termCounts[countKey{term: term, id: id}]
			docCount := index.documentCounts[term]

			tf := 0.5 + 0.5*(scoreNum(termCount)/maxCount)
			idf := scoreNum(math.Log(float64(total / scoreNum(docCount))))

			score += tf * idf
		})

		for pos := 0; pos < Max; pos += 1 {
			if score > scores[pos] {
				copy(scores[pos+1:], scores[pos:Max-1])
				copy(matches[pos+1:], matches[pos:Max-1])

				scores[pos] = score
				matches[pos] = id

				if foundCount < Max {
					foundCount += 1
				}

				break
			}
		}
	}

	return matches[:foundCount]
}
