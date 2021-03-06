package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/go-audio/wav"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var err error

func wavToDSB(ctx *cli.Context, f *os.File, name string) error {
	//w, err := wav.NewReader(f)
	w := wav.NewDecoder(f)
	if err != nil {
		err = errors.Wrap(err, "error occurred while processing .wav file")
		return cli.NewExitError(err, 3)
	}

	w.ReadInfo()
	ch := int(w.NumChans)
	byteRate := int(w.BitDepth/8) * ch
	bps := byteRate / ch
	fs := int(w.SampleRate)

	if fs != 48000 || bps != 2 || w.WavAudioFormat != 1 {
		errMsg := fmt.Sprintf(`audio format error: wav format must be as follows.
sample rate: want 48000 Hz, got %v Hz
sampling bit rate: want 16 bits per sample, got %v bits per sample
audio format: want 1 (PCM), got %v`, fs, w.BitDepth, w.WavAudioFormat)
		return cli.NewExitError(errMsg, 3)
	}

	//data, err := w.ReadSamples(int(w.GetSubChunkSize()) / byteRate * ch)
	aBuf, err := w.FullPCMBuffer()
	if err != nil {
		err = errors.Wrap(err, "error occurred while processing .wav file")
		return cli.NewExitError(err, 3)
	}

	//value, ok := data.([]int16)
	if aBuf.SourceBitDepth != 16 {
		err = errors.New("sampling bit rate is incorrect. need 16 bits per sample")
		err = errors.Wrap(err, "error occurred while processing .wav file")
		return cli.NewExitError(err, 3)
	}

	//iter := len(value) / ch
	iter := aBuf.NumFrames() / ch
	wBuf := make([]byte, iter*bps)

	var fw *os.File
	for j := 0; j < ch; j++ {
		if ch == 1 {
			fw, err = os.Create(fmt.Sprintf("%s.DSB", name))
		} else {
			fw, err = os.Create(fmt.Sprintf("%s_ch%d.DSB", name, j+1))
		}
		if err != nil {
			return cli.NewExitError(err, 3)
		}
		defer fw.Close()

		for i := 0; i < iter; i++ {
			fmt.Printf("working... %d%%\r", (i+1)*100/iter)
			b := new(bytes.Buffer)
			//err = binary.Write(b, binary.LittleEndian, value[ch*i+j])
			err = binary.Write(b, binary.LittleEndian, int16(aBuf.Data[ch*i+j]))
			if err != nil {
				err = errors.Wrap(err, "internal error: error occurred while writing data to buffer")
				return cli.NewExitError(err, 5)
			}
			copy(wBuf[bps*i:bps*(i+1)], b.Bytes())
		}
		_, err = fw.Write(wBuf)
		if err != nil {
			err = errors.Wrap(err, "error occurred while writing data to .DSB file")
			return cli.NewExitError(err, 3)
		}

	}
	if ch == 1 {
		fmt.Printf("\n\n%d file created as %s.DSB\n", ch, name)
	} else {
		fmt.Printf("\n\n%d files created as %s_chX.DSB\n", ch, name)
	}

	fmt.Printf("\n")
	fmt.Println("end!!")

	return nil
}
