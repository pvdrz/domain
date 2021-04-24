use crate::document::Document;

use serde::{Deserialize, Serialize};

use std::fs::File;

#[derive(Serialize, Deserialize)]
#[repr(transparent)]
struct Backup {
    docs: Vec<Document>,
}

pub fn load(path: &str) -> std::io::Result<Vec<Document>> {
    let file = File::open(path)?;
    let backup: Backup = serde_json::from_reader(file)?;

    Ok(backup.docs)
}

pub fn dump(path: &str, docs: Vec<Document>) -> std::io::Result<()> {
    let file = File::create(path)?;
    serde_json::to_writer_pretty(file, &Backup { docs })?;

    Ok(())
}
