# RA2-Style Sprite Assets (Placeholder Directory)

## Status
Currently using improved 3D models with RA2-inspired color palette. This directory is reserved for future 2D sprite integration.

## Potential Free Sources
- **The Spriters Resource**: https://www.spriters-resource.com/pc_computer/commandconquerredalert2/ (has Allied/Soviet building sprites - copyright applies, reference only)
- **OpenRA Project**: https://github.com/OpenRA/OpenRA - Uses original game assets (requires owning the game)
- **OpenGameArt.org**: Search for "isometric military" or "RTS buildings"
- **Kenney.nl**: Free isometric asset packs (not military-themed but good base)
- **itch.io**: Search "isometric RTS sprites" for free/CC0 packs

## Integration Plan
When sprites are available:
1. Place PNG files in this directory organized by category:
   - `buildings/` - construction_yard.png, power_plant.png, barracks.png, etc.
   - `units/` - mcv.png, tank.png, infantry.png, harvester.png
   - `terrain/` - grass.png, water.png, ore.png, etc.
   - `effects/` - explosion.png, muzzle.png
2. Update `engine/render3d/renderer3d.go` to render sprites as textured billboards
3. Map sprite filenames to entity BuildingName keys
