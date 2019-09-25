#!/bin/bash

echo "构建html页面,请等待..."

npm run build

echo "构建可运行程序打包,请等待..."

astilectron-bundler -v

PKG_PATH="output/darwin-amd64/RDM.app/Contents/MacOS/"

RES_PATH="resources"

cp ${RES_PATH} ${PKG_PATH}

echo "打包已完成"

open "output/darwin-amd64/RDM.app"
