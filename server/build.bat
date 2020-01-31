@ECHO off

rd /s/q output

ECHO "build html, waiting..."
cd ..

npm run build

ECHO "build app..."

cd server/resources

RENAME dist app

cd ..

astilectron-bundler -v
