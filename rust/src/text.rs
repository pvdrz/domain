use std::collections::BTreeMap;

use crate::document::Document;
use crate::storage::DocumentId;

type Score = f32;
type Count = u32;

#[derive(PartialEq, Eq, PartialOrd, Ord, Clone, Copy)]
#[repr(transparent)]
struct Gram<const N: usize>([u8; N]);

impl<const N: usize> Gram<N> {
    fn iter_slice<'a>(slice: &'a [u8]) -> impl Iterator<Item = Self> + 'a {
        slice.windows(N).map(|bytes| {
            let mut bytes = unsafe { *(bytes as *const [u8] as *const [u8; N]) };
            bytes.make_ascii_lowercase();
            Self(bytes)
        })
    }
}

pub struct Index<const N: usize> {
    /// How many times a gram appears in a document.
    gram_counts: BTreeMap<(Gram<N>, DocumentId), Count>,
    /// How many documents have a gram.
    document_counts: BTreeMap<Gram<N>, Count>,
    /// How many times the most common gram appears in a document.
    max_gram_counts: BTreeMap<DocumentId, Count>,
}

impl<const N: usize> Index<N> {
    pub fn new() -> Self {
        Self {
            gram_counts: BTreeMap::new(),
            document_counts: BTreeMap::new(),
            max_gram_counts: BTreeMap::new(),
        }
    }

    pub fn search<const MAX: usize>(&self, query: &[u8]) -> Vec<(DocumentId, Score)> {
        assert_ne!(0, MAX);

        let total = self.max_gram_counts.len() as Score;
        let mut scores = Vec::<(DocumentId, Score)>::with_capacity(MAX + 1);

        for (id, max_count) in self.max_gram_counts.iter() {
            let mut score = Score::default();

            let max_count = *max_count as Score;

            for gram in Gram::iter_slice(query) {
                let gram_count = self
                    .gram_counts
                    .get(&(gram, *id))
                    .copied()
                    .unwrap_or_default();

                let document_count = self.document_counts.get(&gram).copied().unwrap_or_default();

                let tf = 0.5 + 0.5 * (gram_count as Score / max_count);

                let idf = (total / document_count as Score).ln();

                score += tf * idf;
            }

            match scores.binary_search_by(|(_, s1)| score.partial_cmp(s1).unwrap()) {
                Ok(index) | Err(index) => {
                    scores.insert(index, (*id, score));
                    if scores.len() > MAX {
                        scores.pop().unwrap();
                    }
                }
            }
        }

        scores
    }

    pub fn insert(&mut self, id: DocumentId, document: &Document) {
        let mut count = BTreeMap::new();

        let grams = Gram::iter_slice(document.title.as_bytes())
            .chain(
                document
                    .authors
                    .iter()
                    .flat_map(|author| Gram::iter_slice(author.as_bytes())),
            )
            .chain(
                document
                    .keywords
                    .iter()
                    .flat_map(|keyword| Gram::iter_slice(keyword.as_bytes())),
            );

        for gram in grams {
            *count.entry(gram).or_default() += 1;
        }

        let mut max_count = 0;

        for (gram, count) in count {
            self.gram_counts.insert((gram, id), count);

            *self.document_counts.entry(gram).or_default() += 1;

            if count > max_count {
                max_count = count;
            }
        }

        self.max_gram_counts.insert(id, max_count);
    }

    pub fn remove(&mut self, id: DocumentId) {
        self.max_gram_counts.remove(&id);
    }
}
