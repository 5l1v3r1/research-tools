package tools

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/mjibson/go-dsp/fft"
	"gonum.org/v1/gonum/floats"
)

func ReadDataFromCSV(inputPath string) (ds []float64, ys []float64, es []float64) {
	fr, err := os.Open(inputPath)
	check(err)
	sc := bufio.NewScanner(fr)
	var ss []string
	var d float64
	var y float64
	var e float64
	for sc.Scan() {
		ss = strings.Split(sc.Text(), ",")
		d, err = strconv.ParseFloat(ss[0], 64)
		check(err)
		ds = append(ds, d)
		y, err = strconv.ParseFloat(ss[1], 64)
		check(err)
		ys = append(ys, y)
		e, err = strconv.ParseFloat(ss[2], 64)
		check(err)
		es = append(es, e)
	}
	return ds, ys, es
}

func ReadCoefFromCSV(inputPath string) (ws []float64) {
	fr, err := os.Open(inputPath)
	check(err)
	sc := bufio.NewScanner(fr)
	var ss []string
	var w float64
	for sc.Scan() {
		ss = strings.Split(sc.Text(), ",")
		w, err = strconv.ParseFloat(ss[0], 64)
		check(err)
		ws = append(ws, w)
	}
	return ws
}

func ReadDataFromWav(name string) []int {
	f, err := os.Open(name)
	check(err)
	defer f.Close()
	wavFile := wav.NewDecoder(f)
	check(err)

	wavFile.ReadInfo()
	ch := int(wavFile.NumChans)
	//byteRate := int(w.BitDepth/8) * ch
	//bps := byteRate / ch
	fs := int(wavFile.SampleRate)
	fmt.Println("ch", ch, "fs", fs)

	buf, err := wavFile.FullPCMBuffer()
	check(err)
	fmt.Printf("SourceBitDepth: %v\n", buf.SourceBitDepth)

	return buf.Data
}

func SaveDataToWav(data []float64, dataDir string, name string) {
	outputPath := filepath.Join(dataDir, name+".wav")
	fw, err := os.Create(outputPath)
	check(err)

	const (
		SampleRate    = 48000
		BitsPerSample = 16
		NumChannels   = 1
		PCM           = 1
	)

	w1 := wav.NewEncoder(fw, SampleRate, BitsPerSample, NumChannels, PCM)
	aBuf := new(audio.IntBuffer)
	aBuf.Format = &audio.Format{
		NumChannels: NumChannels,
		SampleRate:  SampleRate,
	}
	aBuf.SourceBitDepth = BitsPerSample

	aBuf.Data = Float64sToInts(data)
	err = w1.Write(aBuf)
	check(err)

	err = w1.Close()
	check(err)

	err = fw.Close()
	check(err)

	fmt.Printf("\nwav file saved at: %v\n", outputPath)
}

func NormToMaxInt16(data []float64) []float64 {

	maxAmp := floats.Max(AbsFloat64s(data))
	if maxAmp > math.MaxInt16+1 {
		reductionRate := math.MaxInt16 / maxAmp
		for i, _ := range data {
			data[i] *= reductionRate
		}
	}
	return data
}

func AbsFloat64s(fs []float64) []float64 {
	fsAbs := make([]float64, len(fs))
	for i, v := range fs {
		fsAbs[i] = math.Abs(v)
	}
	return fsAbs
}

func abs(x int) int {
	if x < 0 {
		return -1 * x
	}
	return x
}

func AbsInts(is []int) []int {
	isAbs := make([]int, len(is))
	for i, v := range is {
		isAbs[i] = abs(v)
	}
	return isAbs
}

func LinSpace(start, end float64, n int) []float64 {
	res := make([]float64, n)
	if n == 1 {
		res[0] = end
		return res
	}
	delta := (end - start) / (float64(n) - 1)
	for i := 0; i < n; i++ {
		res[i] = start + (delta * float64(i))
	}
	return res
}

func Float64sToInts(fs []float64) []int {
	is := make([]int, len(fs))
	for i, s := range fs {
		is[i] = int(s)
	}
	return is
}

func IntsToFloat64s(is []int) []float64 {
	fs := make([]float64, len(is))
	for i, s := range is {
		fs[i] = float64(s)
	}
	return fs
}
func Float64sToComplex128s(fs []float64) []complex128 {
	cs := make([]complex128, len(fs))
	for i, f := range fs {
		cs[i] = complex(f, 0)
	}
	return cs
}

func Complex128sToFloat64s(cs []complex128) []float64 {
	fs := make([]float64, len(cs))
	for i, c := range cs {
		fs[i] = real(c)
	}
	return fs
}

func Convolve(xs, ys []float64) []float64 {
	var convLen, sumLen = len(xs), len(ys)
	if convLen > sumLen {
		ys = append(ys, make([]float64, convLen-sumLen)...)
	} else {
		convLen, sumLen = sumLen, convLen
		xs = append(xs, make([]float64, convLen-sumLen)...)
	}
	var rs = make([]float64, convLen)
	var nodeSum float64
	var i, j int
	for i = 0; i < convLen; i++ {
		for j = 0; j < sumLen; j++ {
			if i-j < 0 {
				continue
			}
			nodeSum += xs[i-j] * ys[j]
		}
		rs[i] = nodeSum
		nodeSum = 0
	}
	return rs
}

func ConvolveSame(xs, ys []float64) []float64 {
	var convLen, sumLen = len(xs), len(ys)
	if convLen > sumLen {
		ys = append(ys, make([]float64, convLen-sumLen)...)
	} else {
		convLen, sumLen = sumLen, convLen
		xs = append(xs, make([]float64, convLen-sumLen)...)
	}
	var rs = make([]float64, convLen)
	var nodeSum float64
	var i, j int
	for i = 0; i < convLen; i++ {
		for j = 0; j < sumLen; j++ {
			if i-j < 0 {
				continue
			}
			nodeSum += xs[i-j] * ys[j]
		}
		rs[i] = nodeSum
		nodeSum = 0
	}
	return rs
}

func FastConvolve(xs, ys []float64) []float64 {
	L := len(xs)
	N := len(ys)
	M := N + L - 1

	xsz := append(xs, make([]float64, M-L)...)
	ysz := append(ys, make([]float64, M-N)...)

	var rs = make([]float64, M)
	var Rs = make([]complex128, M)


	fmt.Printf("calcurating fft...\n")

	Xs := fft.FFT(Float64sToComplex128s(xsz))
	Ys := fft.FFT(Float64sToComplex128s(ysz))

	for i := 0; i < M; i++ {
		// progress
		fmt.Printf("calucurating convolution... %d%%\r", (i+1)*100/M)
		Rs[i] = Xs[i] * Ys[i]
	}
	fmt.Printf("\ncalcurating ifft...\n")

	rs = Complex128sToFloat64s(fft.IFFT(Rs))

	return rs
}

func SplitPathAndExt(path string) (string, string) {
	return filepath.Join(filepath.Dir(filepath.Clean(path)), filepath.Base(path[:len(path)-len(filepath.Ext(path))])), filepath.Ext(path)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}