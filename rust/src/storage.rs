use crate::document::Document;

use anyhow::{anyhow, Context, Result};
use sled::{Db, Tree};

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub struct DocumentId(pub [u8; std::mem::size_of::<u64>()]);

impl std::fmt::Display for DocumentId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        u64::from_be_bytes(self.0).fmt(f)
    }
}

pub struct Storage {
    db: Db,
    /// StorageId => Document,
    document_tree: Tree,
    /// Hash => StorageId,
    hash_tree: Tree,
}

impl Storage {
    pub fn open<P: AsRef<std::path::Path>>(path: &P) -> Result<Self> {
        let db = sled::open(path)?;
        let document_tree = db.open_tree("documents")?;
        let hash_tree = db.open_tree("hashes")?;

        Ok(Self {
            db,
            document_tree,
            hash_tree,
        })
    }

    pub fn get(&self, id: DocumentId) -> Result<Document> {
        match self.document_tree.get(id.0)? {
            Some(bytes) => {
                bincode::deserialize::<Document>(&bytes).context("Could not deserialize document.")
            }
            None => Err(anyhow!("Document is not in storage.")),
        }
    }

    pub fn iter(&self) -> impl Iterator<Item = Result<(DocumentId, Document)>> {
        self.document_tree.iter().map(|res| {
            let (key, value) = res?;

            let id = unsafe { *(key.as_ptr() as *const DocumentId) };

            let doc: Document = bincode::deserialize(&value)?;

            Ok((id, doc))
        })
    }

    pub fn insert(&mut self, document: &Document) -> Result<DocumentId> {
        let hash = &document.hash;

        if self.hash_tree.contains_key(hash)? {
            anyhow!(
                "File with hash {} has already been stored.",
                hex::encode(hash)
            );
        }

        let id = self.db.generate_id()?.to_be_bytes();

        self.hash_tree.insert(document.hash, &id)?;

        let bytes = bincode::serialize(document)?;
        self.document_tree.insert(id, bytes)?;

        Ok(DocumentId(id))
    }

    pub fn remove(&mut self, DocumentId(id): DocumentId) -> Result<()> {
        if let Some(bytes) = self.document_tree.remove(id)? {
            let document: Document = bincode::deserialize(&bytes)?;
            self.hash_tree.remove(document.hash)?;
        }

        Ok(())
    }
}
