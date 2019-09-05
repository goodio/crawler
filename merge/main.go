package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	/*path := `E:\data\go\src\github.com\ghaoo\crawler\novel\book\剑来`

	files, _ := filepath.Glob(path + "\\*")
	for _, v := range files {
		fmt.Printf("开始处理文件夹：%s \n", v)
		merge(v)
	}*/

	merge(`E:\data\go\src\github.com\ghaoo\crawler\novel\book\剑来`)
}

func merge(root string) {
	name := filepath.Base(root)

	out_name := filepath.Join(root, name+".txt")

	out_file, err := os.OpenFile(out_name, os.O_CREATE|os.O_WRONLY, 0777)

	if err != nil {
		fmt.Printf("Can not open file %s", out_name)
	}

	bWriter := bufio.NewWriter(out_file)

	bWriter.Write([]byte("## " + name + "\n\n\n"))

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() && !strings.HasSuffix(path, ".txt") {
			fmt.Printf("读取文件：%s \n", info.Name())

			fp, err := os.Open(path)

			if err != nil {
				fmt.Printf("Can not open file %v", err)
				return err
			}

			defer fp.Close()

			bReader := bufio.NewReader(fp)

			for {

				buffer := make([]byte, 1024)
				readCount, err := bReader.Read(buffer)
				if err == io.EOF {
					break
				} else {
					bWriter.Write(buffer[:readCount])
				}

			}

			bWriter.Write([]byte("\n\n"))
		}

		return err
	})

	bWriter.Flush()
}
