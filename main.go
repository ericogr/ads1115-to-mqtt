package main

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	addr           = 0x48
	pointerConv    = 0x00
	pointerConfig  = 0x01
	configSingleA0 = 0xC383 // single-shot, A0, FS=±4.096V, 128SPS
)

func main() {
	fmt.Println("starting...")

	// Inicializa Periph.io
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Abre barramento I²C
	bus, err := i2creg.Open("2") // i2c-2
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	dev := &i2c.Dev{Addr: addr, Bus: bus}

	// Configura ADS1115 para leitura single-shot no canal A0
	buf := []byte{pointerConfig, byte(configSingleA0 >> 8), byte(configSingleA0 & 0xFF)}
	if err := dev.Tx(buf, nil); err != nil {
		log.Fatal(err)
	}

	// Aguarda conversão (~8ms a 128SPS)
	time.Sleep(10 * time.Millisecond)

	// Lê resultado de 2 bytes
	readBuf := make([]byte, 2)
	if err := dev.Tx([]byte{pointerConv}, readBuf); err != nil {
		log.Fatal(err)
	}

	// Converte bytes para valor inteiro
	value := int16(readBuf[0])<<8 | int16(readBuf[1])
	fmt.Printf("Leitura A0: %d\n", value)
}
