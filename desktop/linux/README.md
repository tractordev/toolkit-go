# Linux Native Package

## Install

Before building on Linux, make sure you have the following installed:

```bash
sudo apt install \
  build-essential \
  golang \
  libx11-dev \
  libgtk-3-dev \
  libayatana-appindicator3-dev \
  libwebkit2gtk-4.0-dev
```

## TODO

- Use purego!
- Go through every C.* call and update them to purego (i.e C.CString).
- Remove cgo types
- Test purego without cgo
- Transparency call are succesful but no effect?

- Testing?
  - All api.
  - AppIndicator
