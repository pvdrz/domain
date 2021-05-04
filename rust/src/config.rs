use std::path::PathBuf;

use anyhow::{Context, Result};
use serde::Deserialize;

#[derive(Deserialize)]
pub struct Config {
    pub path: PathBuf,
}

impl Config {
    pub fn load() -> Result<Self> {
        let config_path = dirs::config_dir()
            .expect("Config directory is unknown for this platform.")
            .join("domain")
            .join("config.toml");

        let bytes = std::fs::read(&config_path).with_context(|| {
            format!(
                "Could not read configuration from '{}'.",
                config_path.display()
            )
        })?;

        let mut config: Config = toml::from_slice(&bytes).with_context(|| {
            format!(
                "Could not deserialize configuration from '{}'.",
                config_path.display()
            )
        })?;

        if config.path.is_relative() {
            config.path = dirs::home_dir()
                .expect("Home directory is unknown for this platform.")
                .join(config.path);
        }

        Ok(config)
    }
}
