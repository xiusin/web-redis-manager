#!/usr/local/bin/v run

import term

term.clear()

system("cd frontend && bun run build")

system("cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o rdm && upx -9 rdm && mv rdm ..")

println("done!")
