# Ivan
What's the opposite of a nagging fairy that always gets in your way and never
tells you anything useful? Ivan.

Ivan is a tailor-made, keyboard-driven _Ocarina of Time: Randomizer_ item
tracker, hint tracker, timer, and input viewer.  
Because using a mouse is slow, and you gotta go fast.

# Keyboard input
[![screenshot of ivan](./assets/home-screenshot.png)](./assets/home-screenshot.png)

**Ivan must be focused for keyboard input to work.**

- `Esc` quits the tracker, only works when the timer is stopped (not paused) to
  avoid accidentally closing the tracker.  
  When the timer is not paused, `Esc` will cancel the current input mode.
- `Del` to reset the timer and the tracker, only works when the timer is paused.
- `-` to undo the last action.
- `+` to redo the last undone action.

The state of the tracker is persisted to file in case you close it by mistake
of if someone played _Song of Storms_ nearby. `Del` will reset the tracker
right after launching if needed.

## Item tracker
Basic usage:
1. Press a number on your numpad to select an item zone
2. Press another number to select and upgrade/enable an item

_Item zones_ visually map to your numpad, ie. bottom-left is `1`, top-right is
`9`. same for items inside their own 3×3 grid.  
eg. if you wanted to enable _Light Arrows_, you would press `8` (top-middle)
then `2` (bottom-middle). That's it.

Songs are a special case as they are not selectable using their visible
position on the tracker, instead they are accessible in logical order (eg. to
get _Requiem_ you would press `3` as teleportation songs are in the bottom-right
zone, then `4` because _Requiem_ is the fourth song in the pause menu).

Other keys:
- `0` to display the region highlight or cancel your selection.
- `Esc` to cancel your selection.
- `.` to _downgrade_ the next selected item instead of upgrading it.
- `-` to undo the last action.
- `+` to redo the last undone action.

### Mouse
1. Left click to _upgrade_ an item.
2. Right click to _downgrade_ an item.
3. Scroll up/down to:
  - _upgrade_ or _downgrade_ an item.
  - cycle up/down the list of dungeons on stones and medallions.

## Timer
- `Space` once to start the timer, then to pause/resume it.
- `Del` when it is paused to stop it (and reset all tracker data).

## Hint tracker
1. Press the key corresponding to your hint type (**W**otH, **B**arren, **S**ometimes,
   **A**lways).
2. Type your text.
3. Press `Enter`.

- `w` to enter a _WotH_ Hint (green background, fuzzy location search).
- `g` to enter a _Goal_ Hint (green background, freeform text).
- `b` to enter a _Barren_ Hint (red background, fuzzy location search).
- `s` to enter a _Sometimes_ Hint (blue background, freeform text).
- `a` to enter a _Always_ Hint (yellow background).
- `Esc` to cancel your input.
- `Enter` to submit your input.

As _Always Hints_ have a fixed slot, they get special treatment. The text you input
is parsed as the slot name until the first space, then your text. eg. If you
get _Nocturne of Shadows_ on _Ocarina of Time_ you might press `a` to start the
prompt then `oot = nocturne` then `Enter`.

## Dungeon input
Dungeon input allows you to quickly set which dungeons holds what medallions
when reading the altar at the _Temple of Time_.

1. Press `d` to enter dungeon input mode.
2. Dungeon mode always starts with the _Light Medallion_ and goes in the same
   order as the altar.
3. Press the KP of the medallion that originally holds the dungeon. eg. if
   _Fire Temple_ holds the _Light Medallion_, press the key for _Fire
   Medallion_, ie. `5`.
4. Dungeon mode will advance automatically  to the next medallion and you can
   go back to 3.
5. When the last medallion is entered the three stones are filled randomly and
   dungeon mode is exited automatically.

You can also use `+` and `-` to cycle through medallions to correct a mistake
and `0` to exit.

## Input Viewer
The input viewer displays your input around the timer. Button and axes IDs
depend on your configuration and can be set in [config/input_viewer.json](config/input_viewer.json).
If you want to disable the input viewer you can set `Enabled` to `false`.

## Customization
The images in the [`assets`](./assets) folder can be changed if you wish to
customize your background or your icons.
The files in the [config directory](./config) contain all items, keybinds,
layouts, locations, hints, etc.
