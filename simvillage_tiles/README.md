# SimVillage_Tiles

This package contains a very simple tile based renderer and will hopefully be a nifty village simulation soon.

## Features

* Simple tile based renderer
* Player controlled character
* NPC characters (rudimentary)
* Drawable prefab objects (rudimentary)
* Collision detection (rudimentary)
* Chunk loading / chunk generation
* Chunk caching (sorta)

## TODO

* Decouple chunk size from viewport size (?)
* Generation or loading of larger maps
* [WIP] Better layer system / named layers
  * [DONE] Create new structs for handling map chunks and layers
  * [DONE] Migrate renderer to MapChunk and Layer types
  * Allow arbitrary layer names
  * Allow enabling per-layer collision detection via layer property
* Per-Tile actions / events (doors, triggers, ...)
* Indoor maps
* Improve tile render order
* Objects / resources / etc.
* AI for creatures
* Persistent world (since we use procgen, this'll be interesting)

![alt text](https://raw.githubusercontent.com/Flokey82/go_gens/master/simvillage_tiles/images/rgb.png "Screenshot!")