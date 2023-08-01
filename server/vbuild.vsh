#!/usr/local/bin/v run

import term

term.clear()

system('go build -ldflags "-H windowsgui -s -w" -o rdm.exe')

system('upx -9 rdm.exe')

println("done!")
