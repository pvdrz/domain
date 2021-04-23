use crate::document::Document;

use serde::{Deserialize, Serialize};

use std::fs::File;

#[derive(Serialize, Deserialize)]
struct OldBackup {
    docs: std::collections::BTreeMap<usize, Document>,
}

#[derive(Serialize, Deserialize)]
#[repr(transparent)]
struct NewBackup {
    docs: Vec<Document>,
}

pub fn load(path: &str) -> std::io::Result<Vec<Document>> {
    let file = File::open(path)?;
    let backup: OldBackup = serde_json::from_reader(file)?;

    Ok(backup.docs.into_iter().map(|(_, value)| value).collect())
}

pub fn dump(path: &str, documents: &Vec<Document>) -> std::io::Result<()> {
    let file = File::create(path)?;
    let backup = unsafe { &*(documents as *const Vec<Document> as *const NewBackup) };
    serde_json::to_writer_pretty(file, backup)?;

    Ok(())
}
