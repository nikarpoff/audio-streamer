@echo off
set PATH=V:\portaudio\bin;C:\mingw64\bin;%PATH%

set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

set CGO_CFLAGS=-IV:/portaudio/include -D_WIN32_WINNT=0x0600
set CGO_LDFLAGS=-LV:/portaudio/lib -lportaudio -lwinmm -lole32 -luuid -lksuser

set PKG_CONFIG=

go build -tags portaudio -o client.exe

if exist "dist" rmdir /s /q dist
mkdir dist
copy client.exe dist\
rm client.exe
copy "V:\portaudio\bin\libportaudio.dll" dist\
copy "C:\mingw64\bin\libstdc++-6.dll" dist\
copy "C:\mingw64\bin\libgcc_s_seh-1.dll" dist\
copy "C:\mingw64\bin\libwinpthread-1.dll" dist\