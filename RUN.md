# Running & Building Conduit

Conduit is a **native desktop application** built in pure Go with the
[Fyne](https://fyne.io) toolkit. Fyne renders through OpenGL, so building it
requires a C toolchain and a few system graphics/development libraries. This
page lists exactly what to install per platform, then how to build and run.

> **TL;DR (Ubuntu/Debian):**
> ```bash
> sudo apt-get update
> sudo apt-get install -y golang gcc libgl1-mesa-dev xorg-dev
> git clone https://github.com/geekpratyush/conduit.git
> cd conduit
> go run ./cmd/conduit
> ```

---

## 1. Prerequisites

### Common (all platforms)

| Requirement | Notes |
|-------------|-------|
| **Go 1.26+** | `go version` should report 1.26 or newer. |
| **A C compiler** | Fyne uses cgo. GCC/Clang on Linux/macOS, MinGW-w64 or MSVC on Windows. |
| **A graphical display** | Conduit is a GUI, not a headless service. |

### Linux — required system libraries

Fyne needs the OpenGL and X11 development headers. Install the set for your
distribution:

**Debian / Ubuntu / Mint / Pop!_OS**
```bash
sudo apt-get update
sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev
```

**Fedora / RHEL / CentOS**
```bash
sudo dnf install -y gcc libX11-devel libXcursor-devel libXrandr-devel \
    libXinerama-devel libXi-devel mesa-libGL-devel
```

**Arch / Manjaro**
```bash
sudo pacman -S --needed gcc libx11 libxcursor libxrandr libxinerama libxi mesa
```

**openSUSE**
```bash
sudo zypper install -y gcc libX11-devel libXcursor-devel libXrandr-devel \
    libXinerama-devel libXi-devel Mesa-libGL-devel
```

> **Why these?** `libgl1-mesa-dev` provides the OpenGL headers (`gl.pc`), and
> `xorg-dev` pulls in the X11 development packages including **Xcursor**,
> **Xrandr**, **Xinerama**, and **Xi** that Fyne links against. Without them the
> build fails with errors such as `Package gl was not found` or
> `Package xcursor was not found`.

**Verify the headers are present** before building:
```bash
pkg-config --exists gl && echo "gl OK"       || echo "gl MISSING"
pkg-config --exists xcursor && echo "xcursor OK" || echo "xcursor MISSING"
```
Both must print `OK`.

**Wayland users:** Conduit runs fine under Wayland via XWayland (the default). A
native Wayland build is possible with `-tags wayland` but is not required.

### macOS

```bash
xcode-select --install     # provides the C toolchain + OpenGL frameworks
```
Nothing else is needed — the OpenGL framework ships with macOS.

### Windows

- Install a C toolchain — the simplest is **MSYS2 / MinGW-w64** (`gcc`).
- Ensure `gcc` is on your `PATH`.
- No separate OpenGL package is required (provided by the graphics driver).

---

## 2. Build & run

From the repository root:

```bash
# Run directly (compiles + launches)
go run ./cmd/conduit

# Or build a binary
go build -o conduit ./cmd/conduit
./conduit
```

### Run the test suite (no display required)

The core, security, and protocol backends are fully unit-tested and do **not**
need a display or the GUI libraries:

```bash
go test ./internal/...
```

### Package a distributable app (icon + metadata)

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
fyne package -os linux    # or: darwin, windows
```

---

## 2a. Distributing to end users — build-time vs. runtime dependencies

**A pre-built Conduit executable runs on end-user machines without them installing
any of the packages in Section 1.** Those are *development* (`-dev`) packages —
headers and linker stubs needed **only on the machine that compiles** Conduit
(Fyne uses cgo and links OpenGL/X11 at build time). End users never install them.

What a *running* binary needs are the **runtime** shared libraries, which already
ship with any graphical operating system:

| Target OS | Runtime dependency | User must install anything? |
|-----------|--------------------|-----------------------------|
| **macOS** | OpenGL + Cocoa frameworks (part of macOS) | **No** — self-contained `.app`/binary. |
| **Windows** | OpenGL from the graphics driver | **No** — standalone `.exe`. |
| **Linux (desktop)** | `libGL.so.1`, `libX11`, `libXcursor`, `libXrandr`, `libXinerama`, `libXi` — present on every X11/Wayland desktop | **No** — already provided by the desktop. |

So macOS and Windows builds are effectively dependency-free, and a Linux build
runs on any normal desktop because it links against the *runtime* libraries
(e.g. `libgl1`, `libx11-6`), **not** the `-dev` packages.

**Caveats:**

- **One binary per OS.** Because of cgo you build each platform's binary on that
  platform. To produce all three from a single machine, use
  [`fyne-cross`](https://github.com/fyne-io/fyne-cross) (Docker-based).
- **Minimal/headless Linux** (bare containers with no desktop) lack even the
  runtime GL/X11 libraries — but such a host cannot display a GUI anyway. For a
  desktop end user this never applies.

---

## 3. Where Conduit stores data

Conduit keeps per-user state under **`~/.conduit/`**:

| File | Contents |
|------|----------|
| `vault.enc` | AES-256-GCM encrypted credential vault |
| `connections.json` | saved connection profiles (secrets are vault refs, never plaintext) |
| `history.db` | SQLite + FTS5 request/interaction history |

Deleting `~/.conduit/` resets Conduit to a first-run state (samples are re-seeded).

---

## 4. Troubleshooting

| Symptom | Fix |
|---------|-----|
| `Package gl was not found` | Install `libgl1-mesa-dev` (Linux). |
| `Package xcursor was not found` | Install `xorg-dev` (Linux). |
| `cgo: C compiler "gcc" not found` | Install a C toolchain (see prerequisites). |
| Blank window / no display | Ensure `DISPLAY` (X11) or `WAYLAND_DISPLAY` is set; you need a desktop session. |
| `sudo: a password is required` | Run the install command yourself in an interactive shell. |

---

_See [`README.md`](./README.md) for an overview and [`PROJECT_PLAN.md`](./PROJECT_PLAN.md)
for architecture._
