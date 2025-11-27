@echo off
set CGO_CFLAGS=-IF:/vcpkg/installed/x64-windows/include
set CGO_LDFLAGS=-LF:/vcpkg/installed/x64-windows/lib -lportaudio
go build -o audioloop.exe