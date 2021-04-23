mod backup;
mod dbus;
mod document;
mod storage;
mod text;

use std::path::PathBuf;

use document::Document;
use storage::{DocumentId, Storage, StorageResult};
use text::Index;

struct Domain {
    storage: Storage,
    index: Index<3>,
    folder_path: PathBuf,
}

impl Domain {
    fn open(storage_path: &str) -> StorageResult<Self> {
        let storage = Storage::open(storage_path)?;
        let mut index = Index::new();

        for result in storage.iter() {
            let (id, doc) = result?;
            index.insert(id, &doc);
        }

        Ok(Self {
            storage,
            index,
            folder_path: std::env::var_os("DOMAIN_PATH").unwrap().into(),
        })
    }

    fn get(&self, id: DocumentId) -> StorageResult<Option<Document>> {
        self.storage.get(id)
    }

    fn search(&self, query: &str) -> Vec<DocumentId> {
        dbg!(self.index.search::<5>(query.as_bytes()))
            .into_iter()
            .map(|(id, _)| id)
            .collect()
    }

    fn insert(&mut self, document: &Document) -> StorageResult<DocumentId> {
        let id = self.storage.insert(document)?;
        self.index.insert(id, document);

        Ok(id)
    }

    fn remove(&mut self, id: DocumentId) -> StorageResult<()> {
        self.storage.remove(id)?;
        self.index.remove(id);

        Ok(())
    }
}

fn main() {
    let mut domain = Domain::open("test.domain").unwrap();

    for doc in backup::load("/home/christian/MEGAsync/Books/index.json").unwrap() {
        domain.insert(&dbg!(doc)).unwrap();
    }

    dbus::serve(domain).unwrap();
}
