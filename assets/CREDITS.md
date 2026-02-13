# Asset Credits & Licenses

All game assets used in this project are **free** with permissive licenses.

## Kenney.nl Assets (CC0 — Public Domain)

All Kenney assets are licensed under [Creative Commons CC0 1.0 Universal](https://creativecommons.org/publicdomain/zero/1.0/).
No attribution required, but credit is given voluntarily.

### Isometric Landscape
- **Source**: https://kenney.nl/assets/isometric-landscape
- **Used for**: Terrain tiles (grass, dirt, sand, water, rock, snow, forest, etc.)
- **License**: CC0

### Isometric City  
- **Source**: https://kenney.nl/assets/isometric-city
- **Used for**: Building sprites (barracks, factory, power plant, HQ, turrets, walls)
- **License**: CC0

### Isometric Roads
- **Source**: https://kenney.nl/assets/isometric-roads
- **Used for**: Road, bridge, and concrete terrain tiles
- **License**: CC0

### Tanks
- **Source**: https://kenney.nl/assets/tanks
- **Used for**: Tank, harvester, MCV, and vehicle unit sprites
- **License**: CC0

### Smoke Particles
- **Source**: https://kenney.nl/assets/smoke-particles
- **Used for**: Explosion, smoke, and muzzle flash effects
- **License**: CC0

### Particle Pack
- **Source**: https://kenney.nl/assets/particle-pack
- **Used for**: Ore sparkle and additional particle effects
- **License**: CC0

### UI Pack
- **Source**: https://kenney.nl/assets/ui-pack
- **Used for**: UI buttons, panels, and interface elements (reserved for future use)
- **License**: CC0

## Processing

Assets were downloaded and processed using a custom Go tool (`tools/process_assets/main.go`):
- Resized to match the engine's isometric grid (128×64 base tiles)
- Faction-tinted variants generated for Allied (blue) and Soviet (red) teams
- Construction stage and damaged variants generated programmatically
- Tank sprites rotated to generate 8-directional movement frames
- Infantry units generated procedurally

## Author

Original assets by [Kenney](https://kenney.nl) — thank you for making amazing free game assets! 🎮
