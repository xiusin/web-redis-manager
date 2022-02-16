#!/usr/local/bin/v run

import term

term.clear()

// println ('编译中...')

// res := exec("yarn build") or {
// 	panic(err)
// }
// if res.exit_code != 0 {
//     eprintln(res.output)
//     exit(1)
// }

// rmdir_all("../goqt-redis/dist/")

// mkdir("../goqt-redis/dist/")?

// cp_all("server/resources/app", "../goqt-redis/dist/", true) or {
//     println("dist: ${term.fail_message(err)}")
//     return
// }

// system('CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o rdm')
system('go build -ldflags "-H windowsgui -s -w" -o rdm.exe')

system('upx -9 rdm.exe')

println("done!")
