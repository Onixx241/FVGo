# FVGo
A formal verification frontend for Z3 Solver written in Go

----
<img width="683" height="416" alt="image" src="https://github.com/user-attachments/assets/656cba9c-7ab9-45aa-b5b0-eddc5c26d026" />

## Why?

I was experimenting with Formal Verification for my Arty A7 RTL designs and found that other options were too limited or expensive in terms of licensure, like restricting temporal assertions behind licenses, so I decided to make my own which has been a fun and rewarding experience!

## Usage


1. Download latest `Slang` version at (https://github.com/MikePopoloski/slang)
2. Place the `Slang` exe in the same directory as `FVGo`
3. Place desired DUT in the same directory as `FVGo` (Temporary)
4. Run FVGo :

```bash
./FVGo.exe -depth 10 -file='test.sv'
```
`Slang` executable must be in the same directory as FVGo (For now) PATH-ing coming soon


# Status
- **Assert:** Simple temporal assertions, Immediate not supported yet
- **Assume:** Simple Assume properties
- **Cover:** No Cover properties yet
- **Language Support:** Limited support for more advanced SystemVerilog features

## Goals

- Full IEEE 1800-2023 compliance for complete SystemVerilog support
- Utilization of Go Concurrency for better performance as `FVGo` grows
- Multiple DUT support for verifying many designs in one Go
- Full GUI in the future, maybe



## License

This project is licensed under the **GNU General Public License v3.0 (GPL-3.0)**.
