use std::path::PathBuf;

use serde::Deserialize;

#[derive(Debug)]
pub enum ConfigError {
    Toml(toml::de::Error),
    Io(std::io::Error),
}

impl From<toml::de::Error> for ConfigError {
    fn from(err: toml::de::Error) -> Self {
        Self::Toml(err)
    }
}

impl From<std::io::Error> for ConfigError {
    fn from(err: std::io::Error) -> Self {
        Self::Io(err)
    }
}

#[derive(Deserialize)]
pub struct Config {
    pub path: PathBuf,
}

impl Config {
    pub fn load() -> Result<Self, ConfigError> {
        let config_path = dirs::config_dir()
            .unwrap()
            .join("domain")
            .join("config.toml");

        let bytes = std::fs::read(config_path)?;
        let mut config: Config = toml::from_slice(&bytes)?;

        if config.path.is_relative() {
            config.path = dirs::home_dir().unwrap().join(config.path);
        }

        Ok(config)
    }
}
