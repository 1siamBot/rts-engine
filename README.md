# ⚔️ RTS Engine

A Real-Time Strategy game engine written in **Go**, inspired by Command & Conquer: Red Alert 2.

## Features

### Phase 1 — Foundation ✅
- Isometric 2D tile renderer with terrain colors
- Camera system: pan (WASD/edge scroll), zoom (scroll wheel)
- 64x64 tile map with varied terrain (grass, water, forest, ore, roads, cliffs)
- Entity Component System (ECS) architecture
- Deterministic fixed-timestep game loop (20 ticks/sec)
- Unit spawning, selection (click), movement (right-click)
- Minimap with camera viewport indicator
- HUD overlay with game info
- Player system with factions and resources
- Event bus for game events

### Implemented ✅
- **Phase 2**: Pathfinding (A*, Flowfield, NavGrid, Steering)
- **Phase 3**: Combat system (weapons, projectiles, damage types, armor)
- **Phase 4**: Economy (harvesting, building placement, tech tree)
- **Phase 5**: AI system (multi-difficulty)
- **Phase 6**: Map Editor (terrain painting, undo/redo, save/load)
- **Phase 7**: UI/HUD (sidebar, build panel, control groups, minimap)
- **Phase 8**: Audio (BGM, SFX framework)
- **Phase 9**: Networking framework (lockstep, replay, lobby)
- **Phase 10**: Fog of War, animations, MCV deploy/undeploy

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.26 |
| Graphics | Ebitengine v2 |
| Architecture | ECS (Entity Component System) |
| Networking | Custom UDP lockstep (planned) |
| Scripting | GopherLua (planned) |
| Platforms | PC, Web (WASM), Mobile |

## Quick Start

```bash
# Clone
git clone https://github.com/1siamBot/rts-engine.git
cd rts-engine

# Build & Run
go build -o rts-game ./cmd/game/
./rts-game
```

## Controls

| Key | Action |
|-----|--------|
| WASD / Arrow Keys | Pan camera |
| Mouse Scroll | Zoom in/out |
| Middle Mouse Drag | Pan camera |
| Left Click | Select unit |
| Right Click | Move selected units |
| G | Toggle grid overlay |
| M | Toggle minimap |
| Shift + Click | Add to selection |

## Architecture

```
engine/
├── core/       # ECS, GameLoop, Events, Player
├── render/     # Isometric renderer, Camera
├── maplib/     # Tile map system
├── input/      # Input handling
├── pathfind/   # A*, Flowfield (Phase 2)
├── physics/    # Collision (Phase 3)
├── ai/         # AI behaviors (Phase 5)
├── network/    # Multiplayer (Phase 9)
├── audio/      # Sound system (Phase 8)
├── ui/         # HUD, menus (Phase 7)
├── script/     # Lua scripting (Phase 10)
└── asset/      # Resource loading
```

## Downloads

Pre-built binaries for all platforms are available on the [GitHub Releases](https://github.com/1siamBot/rts-engine/releases) page.

| Platform | Game | Map Editor |
|----------|------|------------|
| macOS Apple Silicon | `rts-game-darwin-arm64` | `rts-editor-darwin-arm64` |
| macOS Intel | `rts-game-darwin-amd64` | `rts-editor-darwin-amd64` |
| Linux x64 | `rts-game-linux-amd64` | `rts-editor-linux-amd64` |
| Linux ARM64 | `rts-game-linux-arm64` | `rts-editor-linux-arm64` |
| Windows x64 | `rts-game-windows-amd64.exe` | `rts-editor-windows-amd64.exe` |
| Web (WASM) | `rts-engine-web.zip` | — |

### macOS
```bash
chmod +x rts-game-darwin-arm64
./rts-game-darwin-arm64
```

### Linux
```bash
chmod +x rts-game-linux-amd64
./rts-game-linux-amd64
```

### Windows
Double-click `rts-game-windows-amd64.exe` or run from terminal.

### Web (WASM)
Download `rts-engine-web.zip`, extract, and serve with any HTTP server:
```bash
unzip rts-engine-web.zip
python3 -m http.server 8080
# Open http://localhost:8080
```

### Headless Mode
On systems without a display (e.g., headless servers):
```bash
./rts-game --screenshot output.png   # Render one frame to PNG
./rts-game --headless                 # Same as --screenshot screenshot.png
```

## License

MIT
