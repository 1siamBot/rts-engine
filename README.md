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

### Planned
- **Phase 2**: Pathfinding (A*, Flowfield)
- **Phase 3**: Combat system (weapons, projectiles, AoE)
- **Phase 4**: Economy (harvesting, building, tech tree)
- **Phase 5**: AI system
- **Phase 6**: Map Editor
- **Phase 7**: UI/HUD (sidebar, build panel)
- **Phase 8**: Audio (BGM, SFX)
- **Phase 9**: Multiplayer (lockstep deterministic)
- **Phase 10**: Polish (Fog of War, animations, campaigns)

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

## License

Private — © One Siam Soft Co., Ltd.
