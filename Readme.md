<!--
SPDX-License-Identifier: MIT
-->

# m1ddctui

m1ddctui offers a simple text based user interface for controlling the
brightness of contrast of the default monitor.

## Pre-requisites:

1. Currently, the tool uses [m1ddc](https://github.com/waydabber/m1ddc)
   underneath to interact with the connected monitor.

2. Hence, this works only on macs with M1 processor and monitors that support DDC.

3. Also, the DDC feature should be enabled on the monitor.

## Installation

1. Install [m1ddc](https://github.com/waydabber/m1ddc) using the steps
   mentioned in that respository.

2. Install m1ddctui tool using go command: 
   `go install github.com/manoranjith/m1ddctui`.

3. Create a config file (you can use presets.yaml) in `$HOME/.config/m1ddctui` directory.

4. Run this command to open the application: `m1ddctui`.


## LICENSE

This tool is licensed under MIT license.


