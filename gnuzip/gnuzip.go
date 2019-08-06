package gnuzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
)

const BufferSize = 1024

/*
 * Compress all files given in 'files'
 * the files must be absolute path
 */
func Compress(zip, file string) error {
	fout, err := os.OpenFile(zip, os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return fmt.Errorf("unable to open new gzip file: %v", err)
	}
	defer fout.Close()

	zw := gzip.NewWriter(fout)

	filename := path.Base(file)
	zw.Name = filename

	//- Open the file to compress
	fin, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("unable to open input file: %v", err)
	}
	defer fin.Close()

	//- Compress file. The loop allow to handle chunck of data to to use a more memory-conservative approach
	br := make([]byte, BufferSize)
	for {
		//- Read chunck of input data
		i, err := fin.Read(br)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error while reading input file: %v", err)
		}

		//- Write the chunck of data into the output gzip file
		br = br[:i]
		_, err = zw.Write(br)
		if err != nil {
			return fmt.Errorf("error while writing data into gzip file: %v", err)
		}
	}
	zw.Close()
	return nil
}
