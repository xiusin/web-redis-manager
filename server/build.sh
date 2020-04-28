#!/bin/bash

rm -rf output

echo "构建html页面,请等待..."
cd ..
npm run build
echo "构建可运行程序打包,请等待..."

cd server

astilectron-bundler
