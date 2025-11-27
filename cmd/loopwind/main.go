package main

import (
	"fmt"
	"log"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate      = 44100
	framesPerBuffer = 256
	channels        = 2
)

func main() {
	// Инициализация PortAudio
	if err := portaudio.Initialize(); err != nil {
		log.Fatal("Ошибка инициализации PortAudio:", err)
	}
	defer portaudio.Terminate()

	// Получение списка устройств
	devices, err := portaudio.Devices()
	if err != nil {
		log.Fatal("Ошибка получения устройств:", err)
	}

	// Поиск ASIO устройств
	// var asioDevice *portaudio.DeviceInfo
	// for _, device := range devices {
	// 	if device.HostApi.Type == portaudio.HostApiType(portaudio.ASIO) {
	// 		fmt.Printf("ASIO устройство: %s\n", device.Name)
	// 		asioDevice = device
	// 		break
	// 	} else {
	// 		fmt.Printf("%s Устройство: %s\n", device.HostApi.Type, device.Name)
	// 	}
	// }

	// if asioDevice == nil {
	// 	// log.Fatal("ASIO устройства не найдены")
	// 	asioDevice = devices[2]
	// }

	// Настройка параметров потока
	streamParams := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   devices[3],
			Channels: 1,
			Latency:  devices[3].DefaultLowInputLatency,
		},
		Output: portaudio.StreamDeviceParameters{
			Device:   devices[8],
			Channels: 2,
			Latency:  devices[8].DefaultLowOutputLatency,
		},
		SampleRate:      sampleRate,
		FramesPerBuffer: framesPerBuffer,
	}

	// Создание буферов
	inputBuffer := make([]float32, framesPerBuffer*channels)
	outputBuffer := make([]float32, framesPerBuffer*channels)

	// Создание и открытие потока
	stream, err := portaudio.OpenStream(streamParams, inputBuffer, outputBuffer)
	if err != nil {
		log.Fatal("Ошибка открытия потока:", err)
	}
	defer stream.Close()

	// Запуск потока
	if err := stream.Start(); err != nil {
		log.Fatal("Ошибка запуска потока:", err)
	}

	for {
		// Чтение входных данных
		if err := stream.Read(); err != nil {
			log.Printf("Ошибка чтения: %v", err)
			continue
		}

		// Loopback - копируем вход на выход
		copy(outputBuffer, inputBuffer)

		// Запись выходных данных
		if err := stream.Write(); err != nil {
			log.Printf("Ошибка записи: %v", err)
			continue
		}

		// Небольшая пауза для снижения нагрузки на CPU
		// time.Sleep(1 * time.Millisecond)
	}

	fmt.Println("Запись и воспроизведение запущены... Нажмите Enter для остановки")
	fmt.Scanln()

	if err := stream.Stop(); err != nil {
		log.Fatal("Ошибка остановки потока:", err)
	}
}
