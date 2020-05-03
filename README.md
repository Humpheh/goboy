# GoBoy

[![Build Status](https://travis-ci.org/Humpheh/goboy.svg?branch=master)](https://travis-ci.org/Humpheh/goboy)
[![codecov](https://codecov.io/gh/Humpheh/goboy/branch/master/graph/badge.svg)](https://codecov.io/gh/Humpheh/goboy)
[![Go Report Card](https://goreportcard.com/badge/github.com/Humpheh/goboy)](https://goreportcard.com/report/github.com/Humpheh/goboy)
[![GoDoc](https://godoc.org/github.com/Humpheh/goboy?status.svg)](https://godoc.org/github.com/Humpheh/goboy)

GoBoy is a multi-platform Nintendo GameBoy and GameBoy Color emulator written in go.
The emulator can run the majority of GB games and some CGB games. There is also
colour and sound support.
This emulator was primarily built as a development exercise and is still work in progress.
Please feel free to contribute if you're interested in GameBoy emulator development.

<img src="docs/images/links-awakening-dx.png" width="400"><img src="docs/images/pkmn-gold-game.png" width="400">

The program includes debugging functions making it useful for understanding the emulator operation for
building one yourself. These functions include printing of opcodes and register values to the console at each
step (although will greatly slow down the emulation) and toggling of individual sound channels.

## Installation

Download the [latest release](https://github.com/Humpheh/goboy/releases/latest) of GoBoy from the releases page. 

### Building from source

With go installed, you can install GoBoy into your go bin by running:
```sh
go get github.com/Humpheh/goboy/cmd/goboy
```

If you have Go 1.11 you can also do:
```sh
git clone https://github.com/Humpheh/goboy.git
cd goboy
go build -o goboy cmd/goboy/main.go
```

GoBoy is compatible with MacOS, Windows and Linux. Building on Windows 10 requires MinGW and on Linux, you'll need to install [gtk](https://www.gtk.org/download/linux.php).

GoBoy uses the go library [pixel](https://github.com/faiface/pixel) for control binding and graphics rendering,
which requires OpenGL. You may need to install some requirements which can be found on the
[pixels readme](https://github.com/faiface/pixel#requirements).

## Usage 
```sh
goboy zelda.gb
```
Controls: <kbd>&larr;</kbd> <kbd>&uarr;</kbd> <kbd>&darr;</kbd> <kbd>&rarr;</kbd> <kbd>Z</kbd> <kbd>X</kbd> <kbd>Enter</kbd> <kbd>Backspace</kbd>

The colour palette can be cycled with <kbd>=</kbd> (in DMG mode), and the game can
be made fullscreen with <kbd>F</kbd>.


Other options:
```sh
  -dmg
    	set to force dmg mode
  -mute
    	mute sound output
```

Debug or experimental options:
```sh
  -cpuprofile string
    	write cpu profile to file (debugging)
  -disableVsync
    	set to disable vsync (debugging)
  -stepthrough
    	step through opcodes (debugging)
  -unlocked
    	if to unlock the cpu speed (debugging)
```

### Debugging
There are a few keyboard shortcuts useful for debugging: 

<kbd>Q</kbd> - force toggle background<br/>
<kbd>W</kbd> - force toggle sprites<br/>
<kbd>A</kbd> - print gb background palette data (cgb)<br/>
<kbd>S</kbd> - print sprite palette data (cgb)<br/>
<kbd>D</kbd> - print background map to log<br/>
<kbd>E</kbd> - toggle opcode printing to console (will slow down execution)<br/>
<kbd>7,8,9,0</kbd> - toggle sound channels 1 through 4.

### Saving 
If the loaded rom supports a battery a `<rom-name>.sav` (e.g. `zelda.gb.sav`) file will be created
next to the loaded rom containing a dump of the RAM from the cartridge. A loop in the program will
update this save file every second while the game is running.

## Testing
GoBoy currently passes all of the tests in Blargg's `cpu_instrs` and `instr_timing` test roms.

<img src="docs/images/cpu-instrs.png" width="400"><img src="docs/images/instr-timing.png" width="400">

These roms are included in the source code along with a test to check the output is as expected
(`instructions_test.go` and `timing_test.go`). These tests are also run on each commit.

## Contributing

Please feel free to open pull requests to this project or play around if you're interested! There are
still plenty of small bugs that can easily be found through playing games on the emulator, or take a
task from the TODO list below!

## Known Bugs and TODO list
- [x] Sprites near edge of screen not drawing
- [x] Top half of sprite disappearing off top of screen
- [x] Small sprites row glitch
- [x] BG tile window offset issue - visible on *Pokemon Red* splash screen - possibly mistimed interrupt?
- [x] *Harry Potter and The Chamber of Secrets* has odd sprite issues
- [x] Request to set screen to white does not do so
- [x] MBC3 banking support
- [x] Improve APU (see *Pokemon Yellow* opening screen for reason why)
- [x] Resizable window
- [x] White screen when off
- [x] STOP opcode behaviour
- [ ] Sprite Z-drawing bugs
- [ ] Minor APU timing issues
- [ ] Better APU buffering
- [ ] Stop jittering
- [ ] MBC3 clock support
- [ ] Speed up CPU and PPU
- [ ] Platform native UI?
- [ ] More DMG colour palettes
- [ ] Support save-states
- [ ] Support boot roms
- [ ] [Blargg's test ROMs](http://gbdev.gg8.se/wiki/articles/Test_ROMs)

<img src="docs/images/links-awakening.png" width="400"><img src="docs/images/pkmn-tcg.png" width="400">

<img src="docs/images/pkmn-gold.png" width="400"><img src="docs/images/pkmn-red.png" width="400">

## Resources
A large variety of resources were used to understand and test the GameBoy hardware. Some of these include:
* <http://www.codeslinger.co.uk/pages/projects/gameboy/files/GB.pdf>
* <https://github.com/retrio/gb-test-roms>
* <http://www.codeslinger.co.uk/pages/projects/gameboy/beginning.html>
* <http://bgb.bircd.org/> - invaluable for debugging
* <https://github.com/AntonioND/giibiiadvance/tree/master/docs>
* <https://github.com/trekawek/coffee-gb>
