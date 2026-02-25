@echo off
set PATH=V:\portaudio\bin;C:\mingw64\bin;%PATH%

set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

set CGO_CFLAGS=-IV:/portaudio/include -D_WIN32_WINNT=0x0600
set CGO_LDFLAGS=-LV:/portaudio/lib -lportaudio -lwinmm -lole32 -luuid -lksuser

set PKG_CONFIG_PATH=V:\portaudio\lib\pkgconfig

go build -tags portaudio -o audioloop.exe

copy /Y "V:\portaudio\bin\libportaudio.dll" "libportaudio.dll"