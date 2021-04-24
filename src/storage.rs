use crate::document::Document;

use sled::{Db, Tree};

pub type StorageResult<T> = Result<T, StorageError>;

#[derive(Debug)]
pub enum StorageError {
    Sled(sled::Error),
    Bincode(bincode::Error),
    DuplicatedHash(Box<[u8]>),
}

impl From<sled::Error> for StorageError {
    fn from(err: sled::Error) -> Self {
        Self::Sled(err)
    }
}

impl From<bincode::Error> for StorageError {
    fn from(err: bincode::Error) -> Self {
        Self::Bincode(err)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub struct DocumentId(pub [u8; std::mem::size_of::<u64>()]);

pub struct Storage {
    db: Db,
    /// StorageId => Document,
    document_tree: Tree,
    /// Hash => StorageId,
    hash_tree: Tree,
}

impl Storage {
    pub fn open<P: AsRef<std::path::Path>>(path: &P) -> StorageResult<Self> {
        let db = sled::open(path)?;
        let document_tree = db.open_tree("documents")?;
        let hash_tree = db.open_tree("hashes")?;

        Ok(Self {
            db,
            document_tree,
            hash_tree,
        })
    }

    pub fn get(&self, DocumentId(id): DocumentId) -> StorageResult<Option<Document>> {
        match self.document_tree.get(id)? {
            Some(bytes) => {
                let document: Document = bincode::deserialize(&bytes)?;

                Ok(Some(document))
            }
            None => Ok(None),
        }
    }

    pub fn iter(&self) -> impl Iterator<Item = StorageResult<(DocumentId, Document)>> {
        self.document_tree.iter().map(|res| {
            let (key, value) = res?;

            let id = unsafe { *(key.as_ptr() as *const DocumentId) };

            let doc: Document = bincode::deserialize(&value)?;

            Ok((id, doc))
        })
    }

    pub fn insert(&mut self, document: &Document) -> StorageResult<DocumentId> {
        let hash = &document.hash;

        if self.hash_tree.contains_key(hash)? {
            return Err(StorageError::DuplicatedHash(Box::new(hash.clone())));
        }

        let id = self.db.generate_id()?.to_be_bytes();

        self.hash_tree.insert(document.hash, &id)?;

        let bytes = bincode::serialize(document)?;
        self.document_tree.insert(id, bytes)?;

        Ok(DocumentId(id))
    }

    pub fn remove(&mut self, DocumentId(id): DocumentId) -> StorageResult<()> {
        if let Some(bytes) = self.document_tree.remove(id)? {
            let document: Document = bincode::deserialize(&bytes)?;
            self.hash_tree.remove(document.hash)?;
        }

        Ok(())
    }
}
