use std::collections::HashMap;
use std::convert::TryInto;

use anyhow::{Context, Result};
use hex::FromHex;
use zbus::export::zvariant;
use zbus::{dbus_interface, fdo, Connection, ObjectServer};

use crate::{storage::DocumentId, Domain};

const SERVER_NAME: &str = "com.github.pvdrz.domain";
const SERVER_PATH: &str = "/com/github/pvdrz/domain";

pub(crate) fn serve(domain: Domain) -> Result<()> {
    let connection = Connection::new_session()?;

    log::info!("Requesting server name.");
    fdo::DBusProxy::new(&connection)?
        .request_name(SERVER_NAME, fdo::RequestNameFlags::ReplaceExisting.into())
        .with_context(|| format!("Could not reserve server name '{}'.", SERVER_NAME))?;

    log::info!("Creating server path.");
    let mut object_server = ObjectServer::new(&connection);
    object_server
        .at(
            &SERVER_PATH.try_into().expect("Invalid server path"),
            domain,
        )
        .with_context(|| format!("Could not register server at path '{}'.", SERVER_PATH))?;

    log::info!("Server is up.");
    loop {
        object_server
            .try_handle_next()
            .context("Could not handle message.")?;
    }
}

#[dbus_interface(name = "org.gnome.Shell.SearchProvider2")]
impl Domain {
    fn get_initial_result_set(&self, terms: Vec<&str>) -> Vec<String> {
        let query = terms.join(" ");
        log::info!("Received query \"{}\".", query);
        self.search(&query)
            .into_iter()
            .map(|id| {
                log::info!("Document {} matches query.", id);
                hex::encode(id.0)
            })
            .collect()
    }

    fn get_subsearch_result_set(
        &self,
        _previous_results: Vec<&str>,
        terms: Vec<&str>,
    ) -> Vec<String> {
        self.get_initial_result_set(terms)
    }

    fn get_result_metas(
        &self,
        str_ids: Vec<String>,
    ) -> fdo::Result<Vec<HashMap<&'static str, zvariant::Value>>> {
        let mut metas = Vec::with_capacity(str_ids.len());

        for str_id in str_ids {
            let id = DocumentId(
                <[u8; std::mem::size_of::<u64>()]>::from_hex(&str_id)
                    .context("Client sent an invalid document ID.")
                    .map_err(|e| fdo::Error::Failed(e.to_string()))?,
            );

            let doc = self
                .get(id)
                .map_err(|e| fdo::Error::Failed(e.to_string()))?;

            log::info!("Retrieved metadata for document {}", id);

            let meta = {
                let mut meta = HashMap::with_capacity(3);
                meta.insert("id", str_id.into());
                meta.insert("name", doc.title.into());
                meta.insert("description", doc.authors.join(", ").into());
                meta
            };

            metas.push(meta);
        }

        Ok(metas)
    }

    fn activate_result(&self, str_id: &str, _terms: Vec<&str>, _timestamp: u32) -> fdo::Result<()> {
        let id = DocumentId(
            <[u8; std::mem::size_of::<u64>()]>::from_hex(&str_id)
                .context("Client sent an invalid document ID.")
                .map_err(|e| fdo::Error::Failed(e.to_string()))?,
        );

        let doc = self
            .get(id)
            .map_err(|e| fdo::Error::Failed(e.to_string()))?;

        let path = self
            .config
            .path
            .join(hex::encode(doc.hash))
            .with_extension(doc.extension);

        log::info!("Opening path \"{}\".", path.display());

        open::that(&path)
            .map(|_| ())
            .with_context(|| format!("Failed to open document at path '{}'", path.display()))
            .map_err(|e| fdo::Error::Failed(e.to_string()))
    }

    fn launch_search(&self, _terms: Vec<&str>, _timestamp: u32) {}
}
