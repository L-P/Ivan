# Ivan
What's the opposite of a nagging fairy that always gets in your way and never
tells you anything useful? Ivan.

Ivan is a tailor-made, keyboard-driven _Ocarina of Time: Randomizer_ item
tracker, hint tracker (TBD), and timer. Because using a mouse is slow, and you
gotta go fast.

# Binds
[![screenshot of ivan](./assets/home-screenshot.png)](./assets/home-screenshot.png)

- `esc` quits the tracker, only works when the timer is in its zero-state to avoid
  accidentally closing the tracker.

## Item tracker
### Keyboard
**Ivan must be focused for keyboard input to work.**

Basic usage:
1. Press a number to select a region
2. Press another number to upgrade the item

Other keys:
- `0` to display the region highlight or reset your selection.
- `.` to _downgrade_ the next selected item instead of upgrading it.
- `-` to undo the last action.
- `+` to redo the last undone action.

Songs are a special case as they are not selectable using their visible
position on the tracker, instead they are accessible in logical order (ie. to
get Requiem you would press `3` to select the teleportation songs zone then
`4`).

### Mouse
1. Left click to _upgrade_ an item.
2. Right click to _downgrade_ an item.
3. Scroll up/down to:
  - _upgrade_ or _downgrade_ an item.
  - cycle up/down the list of dungeons on stones and medallions.

## Timer
- `space` once to start the timer, then to pause/resume its _display_ (it still
  runs in the background).
- `del` to reset the timer when it's paused.
