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

- Transparency call are succesful but no effect?
- Focus has no effect? What is the intended effet of focus?
- Does purego has static linking, something like linking at compile time and purego just loading them in runtime?
- Reload undoes user effects, like resize and positioning.

- Testing?
  - All api.
  - AppIndicator
