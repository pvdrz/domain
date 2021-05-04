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
use storage::{DocumentId, Storage};
use structopt::StructOpt;
use text::Index;

use anyhow::{Context, Result};

struct Domain {
    storage: Storage,
    index: Index<3>,
    config: Config,
}

impl Domain {
    fn open() -> Result<Self> {
        let config = Config::load()?;
        let storage = Storage::open(&config.path.join("storage/"))?;
        let mut index = Index::new();

        for result in storage.iter() {
            let (id, doc) = result.context("Could not index document.")?;
            index.insert(id, &doc);
        }

        Ok(Self {
            storage,
            index,
            config,
        })
    }

    fn get(&self, id: DocumentId) -> Result<Document> {
        self.storage
            .get(id)
            .with_context(|| format!("Could not get document with ID: {}.", id))
    }

    fn search(&self, query: &str) -> Vec<DocumentId> {
        self.index
            .search::<5>(query.as_bytes())
            .into_iter()
            .map(|(id, _)| id)
            .collect()
    }

    fn insert(&mut self, document: &Document) -> Result<DocumentId> {
        let id = self.storage.insert(document).with_context(|| {
            format!(
                "Could not insert document with hash: {}",
                hex::encode(document.hash)
            )
        })?;
        self.index.insert(id, document);

        Ok(id)
    }

    fn remove(&mut self, id: DocumentId) -> Result<()> {
        self.storage
            .remove(id)
            .with_context(|| format!("Could not remove document with ID: {}.", id))?;
        self.index.remove(id);

        Ok(())
    }

    fn execute(self, command: Command) -> Result<()> {
        match command {
            Command::Serve => dbus::serve(self),
        }
    }
}

fn run() -> Result<()> {
    let command = Command::from_args();
    let domain = Domain::open()?;

    domain.execute(command)
}

fn main() {
    match run() {
        Ok(()) => (),
        Err(err) => eprintln!("{:?}", err),
    }
}
