#!/bin/sh

mkdir -p ~/.config/domain
rm -f ~/.config/domain/config.toml
ln -sf $(pwd)/config.toml ~/.config/domain/config.toml

mkdir -p ~/.config/systemd/user
rm -f ~/.config/systemd/user/domain.service
ln -sf $(pwd)/domain.service ~/.config/systemd/user/domain.service

mkdir -p ~/.local/share/applications
rm -f ~/.local/share/applications/domain.desktop
ln -sf $(pwd)/domain.desktop ~/.local/share/applications/domain.desktop

go build
sudo rm -f /usr/local/bin/domain
sudo ln -sf $(pwd)/domain /usr/local/bin/domain

sudo rm -f /usr/share/gnome-shell/search-providers/com.github.pvdrz.domain.search-provider.ini
sudo ln -sf $(pwd)/com.github.pvdrz.domain.search-provider.ini /usr/share/gnome-shell/search-providers/com.github.pvdrz.domain.search-provider.ini
