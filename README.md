# Go8080

This is a complete and accurate emulator for the Intel 8080 microprocessor. I used the i8080 emulator to implement an emulator for the Space Invaders arcade cabinet. My implementation does not emulate sound. The i8080 emulator is located in `i8080/`, the Space Invaders emulator is located in `i8080Invaders/`, and a test emulator is included in `i8080Test/`.

![](https://github.com/is386/Go8080/blob/main/demo.gif?raw=true)

## Usage

`go run main.go`

This will run the Space Invaders emulator in a separate screen.

## Dependencies

- `go 1.15`

### Go Dependencies

- `github.com/veandco/go-sdl2`

## Testing

`go test -v ./i8080Test/`

This will run a test emulator that runs test ROMs that were used to exercise the instructions on the original Intel 8080. There are four tests ROMs located in `/i8080Test/roms`. Currently, my i8080 emulator can pass all of the test ROMs.

## Space Invaders Controls

|   Key   |       Effect        |
| :-----: | :-----------------: |
| `SPACE` |     Insert Coin     |
|   `1`   | Start 1 Player Game |
|   `2`   | Start 2 Player Game |
|   `A`   |      Move Left      |
|   `D`   |     Move Right      |
|   `J`   |        Shoot        |
