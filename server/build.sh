#!/bin/bash

rm -rf output

echo "构建html页面,请等待..."
cd ..
npm run build
echo "构建可运行程序打包,请等待..."

cd server

astilectron-bundler -v

echo "打包已完成, 打开程序"

open "output/darwin-amd64/RedisManager.app"
