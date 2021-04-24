use std::collections::HashMap;
use std::convert::TryInto;

use hex::FromHex;
use zbus::export::zvariant;
use zbus::{dbus_interface, fdo, Connection, ObjectServer};

use crate::{storage::DocumentId, Domain};

const SERVER_NAME: &str = "com.github.pvdrz.domain";
const SERVER_PATH: &str = "/com/github/pvdrz/domain";

pub(crate) fn serve(domain: Domain) -> fdo::Result<()> {
    let connection = Connection::new_session()?;

    fdo::DBusProxy::new(&connection)?
        .request_name(SERVER_NAME, fdo::RequestNameFlags::ReplaceExisting.into())?;

    let mut object_server = ObjectServer::new(&connection);
    object_server.at(&SERVER_PATH.try_into().unwrap(), domain)?;

    println!("looping");
    loop {
        object_server.try_handle_next()?;
    }
}

#[dbus_interface(name = "org.gnome.Shell.SearchProvider2")]
impl Domain {
    fn get_initial_result_set(&self, terms: Vec<&str>) -> Vec<String> {
        let query = terms.join(" ");
        self.search(&dbg!(query))
            .into_iter()
            .map(|DocumentId(bytes)| hex::encode(bytes))
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
    ) -> Vec<HashMap<&'static str, zvariant::Value>> {
        let mut metas = Vec::with_capacity(str_ids.len());

        for str_id in str_ids {
            let id = match <[u8; std::mem::size_of::<u64>()]>::from_hex(&str_id) {
                Ok(bytes) => DocumentId(bytes),
                Err(_) => panic!(),
            };

            let doc = match self.get(id) {
                Ok(Some(doc)) => doc,
                Ok(None) | Err(_) => panic!(),
            };

            let meta = {
                let mut meta = HashMap::with_capacity(3);
                meta.insert("id", str_id.into());
                meta.insert("name", doc.title.into());
                meta.insert("description", doc.authors.join(", ").into());
                meta
            };

            metas.push(meta);
        }

        metas
    }

    fn activate_result(&self, str_id: &str, _terms: Vec<&str>, _timestamp: u32) {
        let id = DocumentId(<[u8; std::mem::size_of::<u64>()]>::from_hex(str_id).unwrap());

        let doc = self.get(id).unwrap().unwrap();

        let path = self
            .config
            .path
            .join(hex::encode(doc.hash))
            .with_extension(doc.extension);

        open::that(dbg!(path)).unwrap();
    }

    fn launch_search(&self, _terms: Vec<&str>, _timestamp: u32) {}
}
