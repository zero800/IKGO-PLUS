# IKGO Plus

[![Documentation Status](https://readthedocs.org/projects/ikemen-plus/badge/?version=latest)](https://ikemen-plus.readthedocs.io/en/latest/?badge=latest) [![Go Report Card](https://goreportcard.com/badge/github.com/shinlucho/ikemen-plus)](https://goreportcard.com/report/github.com/shinlucho/ikemen-plus)

Personal modification in fighting game engine based on Mugen.

## New features

1. Add OpenAL32.dll, no need more OpenAL 1.1 Windows Installer

2. Optional, show different animation for Full power bar.

In Fight.def -> [Powerbar] use "p1.fullfront.spr" or "p1.fullfront.anim" for show different Full power bar. 

3. Partial counter for Powerbar.

In Fight.def -> [Powerbar] use p1.partial_bar > 0 for show partial bar (Full every 1000) like kof games. 

4. Play Sound for Maximum Power Level

In Fight.def -> [Powerbar] use 'level5.snd' for play sond when Power = MaxPower (Maximum Level) 

5. Add other PAUSE and STEP shortcut

"Pause" = PAUSE ou F11 / "Step" = SCROLLLOCK or F10

*some notebooks keyboards do not have PAUSE or SCROLL LOCK keys.
