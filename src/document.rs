use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct Document {
    pub title: String,
    pub authors: Vec<String>,
    pub keywords: Vec<String>,
    pub extension: String,
    #[serde(with = "hex")]
    pub hash: [u8; 32],
}

impl Document {
    pub fn new(
        title: String,
        authors: Vec<String>,
        keywords: Vec<String>,
        hash: [u8; 32],
        extension: String,
    ) -> Self {
        Self {
            title,
            authors,
            keywords,
            hash,
            extension,
        }
    }
}
