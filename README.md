# audio-streamer
 Opens WebSocket audio stream between multiple clients


# Installing
### Windows:
1. Clone repository
```
git clone https://github.com/nikarpoff/audio-streamer
```

2. Install portaudio with ASIO capabilities:
2.1. Download ()[ASIO SDK]. Unzip to some package, e.g. C:/ASIOSDK

2.2. Download portaudio-2.0 to some package, e.g. C:/portaudio
```
git clone https://github.com/PortAudio/portaudio
```

2.3. Go to portaudio, create directory build, go to build
```
cd C:/portaudio
mkdir build
cd build
```

2.4. Build portaudio with ASIO capabilities (you can make build.bat) using cmake and mingw64:
```
cmake .. ^
  -G "MinGW Makefiles" ^
  -DCMAKE_INSTALL_PREFIX=C:\portaudio ^
  -DPA_USE_ASIO=ON ^
  -DASIO_SDK_DIR=C:\ASIOSDK
```

2.5. Install portaudio
```
mingw32-make -j8
mingw32-make install
```

3. Go to audio-streamer/cmd/loopwind, change default run.bat or build.bat paths to portaudio and mingw64 to your paths, e.g. C:/portaudio. 

4. Build or run go application
```
cd ./cmd/loopwind
./run.bat
```

## Default server from environment
The client reads `AUDIO_STREAMER_SERVER_ADDR` from process environment.

```
AUDIO_STREAMER_SERVER_ADDR=127.0.0.1:8080
```

If the variable is missing, fallback is `127.0.0.1:8080`.

## CI/CD (GitHub Actions)
On each push to `main`, workflow `.github/workflows/server-cicd.yml`:
1. runs hub test (`internal/network`)
2. verifies `cmd/server` package builds
3. builds Linux server binary
4. copies binary to remote server via SSH
5. restarts service (or runs binary with `nohup` if service name is not provided)

Required GitHub repository secrets:
- `SSH_HOST`
- `SSH_USER`
- `SSH_PRIVATE_KEY`
- `DEPLOY_PATH` (e.g. `/opt/audio-streamer`)

Optional secrets:
- `SSH_PORT` (defaults to `22`)
- `SERVER_SERVICE_NAME` (if using systemd service restart)