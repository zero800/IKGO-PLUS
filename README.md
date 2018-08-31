# Ikemen Plus

[![Documentation Status](https://readthedocs.org/projects/ikemen-plus/badge/?version=latest)](https://ikemen-plus.readthedocs.io/en/latest/?badge=latest) [![Go Report Card](https://goreportcard.com/badge/github.com/shinlucho/ikemen-plus)](https://goreportcard.com/report/github.com/shinlucho/ikemen-plus)

Ikemen Plus is a 2D fighting game engine based on Mugen.

## Building

### Linux

With a debian based system, it can be compiled the following way:

1. Install golang:
`sudo apt install golang-go`

2. Install git:
`sudo apt install git`

3. Install [GLFW](https://github.com/go-gl/glfw) dependencies:
`sudo apt install libgl1-mesa-dev xorg-dev`

4. Install OpenAL dependencies:
`sudo apt install libopenal1 libopenal-dev`

5. Download Ikemen GO Plus repository:
`git clone https://github.com/shinlucho/ikemen-plus.git`

6. Move to downloaded folder:
`cd ikemen-plus`

7. Execute get.sh to download Ikemen dependencies (it takes a while):
`./get.sh`

8. FINALLY compile:
`./build.sh`

9. And now, Ikemen can be opened double clicking Ikemen-GO-Plus, or with the terminal:
`./Ikemen_GO`

