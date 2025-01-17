# Gamerogueish

This is a sorta-kinda roguelike using the fantastic package https://github.com/BigJk/ramen, which is a simple console emulator written in go that can be used to create various ascii / text (roguelike) games.

Right now, the code is a really basic re-factor of the roguelike example that comes with Ramen, but I'll use it as a basis for various (at least to me) interesting experiments :)


![alt text](https://raw.githubusercontent.com/Flokey82/go_gens/master/gamerogueish/images/rgb.png "rogue-ish")

## TODO

* FOV / 'Fog of war'
  * [DONE] Basic radius based FOV
  * Raycasting based FOV
* Creatures
  * [DONE] Basic movement (random)
  * [DONE] AI (basic)
  * Pathfinding
* Documentation
* Inventory
  * [DONE] Basic inventory
  * [DONE] Item add / remove
* Items
  * [DONE] Basic items
  * [DONE] Item generation
  * [DONE] Consumable items
  * [DONE] Equippable items
  * Item pickup / drop
  * Item effects
* Combat
  * Player death
* Map generation
  * [DONE] Custom world generator functions
  * [DONE] Creature placement
  * Neighbor rooms not centered (optionally)
  * Connections / doors not centered (optionally)
  * Caves
  * Custom seed
  * Item placement

## Interesting stuff

* FOV
  * https://github.com/ajhager/rog/blob/master/fov.go
  * http://journal.stuffwithstuff.com/2015/09/07/what-the-hero-sees/
  * http://www.roguebasin.com/index.php?title=Field_of_Vision
* Loot
  * http://journal.stuffwithstuff.com/2014/07/05/dropping-loot/
  * https://www.reddit.com/r/roguelikedev/comments/2y3rkg/faq_friday_7_loot/
* Game loop
  * http://journal.stuffwithstuff.com/2014/07/15/a-turn-based-game-loop/
* Map generation
  * http://journal.stuffwithstuff.com/2014/12/21/rooms-and-mazes/