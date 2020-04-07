package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type params struct {
	*flag.FlagSet `json:"-"`

	inputDir   string
	outputDir  string
	quality    int
	inputFile  string
	outputFile string
	help       bool
}

func main() {
	fmt.Println("欢迎使用go-magick图片压缩工具----甩甩制作")

	p := &params{}
	p.FlagSet = flag.NewFlagSet("go-magick Params", flag.ContinueOnError)

	p.StringVar(&p.inputDir, "inputDir", "", "输入文件夹")
	p.StringVar(&p.outputDir, "outputDir", "", "输出文件夹")
	p.StringVar(&p.inputFile, "inputFile", "", "输入文件名")
	p.StringVar(&p.outputFile, "outputFile", "", "输出文件名")
	p.IntVar(&p.quality, "quality", 80, "压缩质量")
	p.BoolVar(&p.help, "h", false, "帮助信息")

	var err error

	err = p.Parse(os.Args[1:])
	if err != nil {
		p.Usage()
		os.Exit(0)
	}

	if p.help {
		p.Usage()
		return
	}

	if p.inputDir == "" && p.inputFile == "" {
		fmt.Println("输入文件夹和输入文件名称不能同时为空")
		p.Usage()
		return
	}
	if p.inputDir != "" && p.inputFile != "" {
		err = compression(p.inputDir, p.inputFile, p.outputDir, p.outputDir, p.quality)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	wg := new(sync.WaitGroup)

	deepdirsCompression(p.inputDir, p.outputDir, p.quality, wg)

	wg.Wait()
	fmt.Println("恭喜你，已经全部给你ojbk压缩完成")

}

// 深度递归文件
func deepdirsCompression(dirPath string, outputDir string, quality int, wg *sync.WaitGroup) {

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Println("报错了:", err)
		return
	}

	for _, item := range files {
		if item.IsDir() {
			deepdirsCompression(filepath.Join(dirPath, item.Name()), outputDir, quality, wg)
			continue
		}
		fmt.Println("开始压缩:", item.Name())
		wg.Add(1)
		go func(inputDir, inputFileName, outDir string) {
			defer wg.Done()
			err := compression(inputDir, inputFileName, outDir, "", quality)
			if err != nil {
				fmt.Println("压缩出现错误:", err)
				return
			}
			fmt.Println("压缩完成:", inputFileName)

		}(dirPath, item.Name(), outputDir)
	}

}

func compression(inputDir, inputFilename, outdir, outFilename string, quality int) (err error) {
	if outdir == "" {
		outdir, _ = os.Getwd()
	}

	if inputFilename == "" {
		return errors.New("输入文件为空")
	}

	inputFileExt := strings.ToLower(filepath.Ext(inputFilename))

	if inputFileExt != ".png" && inputFileExt != ".jpg" && inputFileExt != ".jpeg" {
		return errors.New("不支持的图片格式：" + inputFileExt)
	}

	if outFilename == "" {
		outFilename = inputFilename
	}
	index := strings.LastIndex(inputFilename, ".")
	if index < 0 {
		return errors.New("请检查文件是否为合法文件")
	}

	preFilename := outFilename[:index-1]
	ext := filepath.Ext(outFilename)
	ext = strings.ToLower(ext)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return errors.New("不支持的图片格式：" + ext)
	}

	filename := filepath.Join(outdir, outFilename)

	i := 0
	for {
		if !fileExists(filename) {
			break
		}
		i++
		filename = filepath.Join(outdir, fmt.Sprintf("%s_out_%d%s", preFilename, i, ext))

	}

	inputFile := filepath.Join(inputDir, inputFilename)

	cmd := exec.Command("./magick.exe", "convert", inputFile, "-quality", fmt.Sprintf("%d", quality), filename)
	return cmd.Run()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
