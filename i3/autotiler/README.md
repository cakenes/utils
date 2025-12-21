### Autotiler for [i3wm](https://i3wm.org/)

Simple autotiler written in <b>GO</b> & <b>BASH</b>, pick your poison.<br />
Will listen for events from <code>i3-msg</code> and compare focused window width & height to pick an split direction.<br />

### Installing (pick one):
- Download & copy over <code>src/i3-autotiler</code> to <code>/usr/local/bin/</code>
- Download <code>PKGBUILD</code> and install with <code>makepkg -si</code>

### Running:
Add <code>exec_always --no-startup-id i3-autotiler</code> to your i3 config.
