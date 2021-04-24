mod backup;
mod command;
mod config;
mod dbus;
mod document;
mod storage;
mod text;

use command::Command;
use config::Config;
use document::Document;
use storage::{DocumentId, Storage, StorageResult};
use structopt::StructOpt;
use text::Index;

struct Domain {
    storage: Storage,
    index: Index<3>,
    config: Config,
}

impl Domain {
    fn open() -> StorageResult<Self> {
        let config = Config::load().unwrap();
        let storage = Storage::open(&config.path.join("storage/"))?;
        let mut index = Index::new();

        for result in storage.iter() {
            let (id, doc) = result?;
            index.insert(id, &doc);
        }

        Ok(Self {
            storage,
            index,
            config,
        })
    }

    fn get(&self, id: DocumentId) -> StorageResult<Option<Document>> {
        self.storage.get(id)
    }

    fn search(&self, query: &str) -> Vec<DocumentId> {
        self.index
            .search::<5>(query.as_bytes())
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

    fn execute(self, command: Command) {
        match command {
            Command::Serve => dbus::serve(self).unwrap(),
        }
    }
}

fn main() {
    let command = Command::from_args();
    let domain = Domain::open().unwrap();

    domain.execute(command)
}
