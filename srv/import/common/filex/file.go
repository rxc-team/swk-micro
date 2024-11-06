package filex

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"rxcsoft.cn/utils/storage"

	"rxcsoft.cn/pit3/srv/import/common/charsetx"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// SaveFile 创建并保存文件到minio
func SaveFile(data []byte, domain, fileName, fileType, contentType, appRoot string) (file *storage.ObjectInfo, e error) {
	err := Mkdir("temp/")
	if err != nil {
		return nil, err
	}

	// 创建临时文件
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	_, err = f.Write(data)
	if err != nil {
		return nil, err
	}

	fo, err := os.Open(fileName)
	defer fo.Close()
	if err != nil {
		return nil, err
	}

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		return nil, err
	}
	filePath := path.Join(appRoot, fileType, fileName)
	result, err := minioClient.SavePublicObject(fo, filePath, contentType)
	if err != nil {
		return nil, err
	}

	/* // 删除临时文件
	err = os.Remove(fileName)
	if err != nil {
		return result, err
	}
	*/
	return result, nil
}

// SaveLocalFile 创建并保存文件本地
func SaveLocalFile(data []byte, fileName string) (e error) {
	// 创建临时文件
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// DeleteFile 删除文件
func DeleteFile(domain, fileName string) {
	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		fmt.Printf("delete file has error :%v", err)
	}
	if err := minioClient.DeleteObject(fileName); err != nil {
		fmt.Printf("delete file has error :%v", err)
	}
}

// WriteAndSaveFile 写入&保存文件
func WriteAndSaveFile(domain, appID string, items []string) (file *storage.ObjectInfo) {
	bytesBuffer := &bytes.Buffer{}
	bytesBuffer.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码

	writer := bufio.NewWriter(bytesBuffer)
	for _, item := range items {
		writer.WriteString(item + "\n")
	}

	writer.Flush() // 此时才会将缓冲区数据写入

	err := Mkdir("temp/")
	if err != nil {
		return nil
	}

	appRoot := "app_" + appID

	filename := "temp/tmp" + "_" + time.Now().Format("20060102150405000") + ".txt"
	f, err := SaveFile(bytesBuffer.Bytes(), domain, filename, "text", "text/plain", appRoot)
	if err != nil {
		return nil
	}

	os.Remove(filename)

	return f
}

// getFileContentType 获取文件类型
func getFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

// Zip 压缩文件夹
func Zip(dir, fileName string) {
	// 预防：旧文件无法覆盖
	os.RemoveAll(fileName)

	// 创建：zip文件
	zipfile, _ := os.Create(fileName)
	defer zipfile.Close()

	// 打开：zip文件
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	// 遍历路径信息
	filepath.Walk(dir, func(path string, info os.FileInfo, _ error) error {

		// 如果是源路径，提前进行下一个遍历
		if path == dir {
			return nil
		}

		// 获取：文件头信息
		header, _ := zip.FileInfoHeader(info)
		header.Name = strings.TrimPrefix(path, dir+`\`)

		// 判断：文件是不是文件夹
		if info.IsDir() {
			header.Name += `/`
		} else {
			// 设置：zip的文件压缩算法
			header.Method = zip.Deflate
		}

		// 创建：压缩包头部信息
		writer, _ := archive.CreateHeader(header)
		if !info.IsDir() {
			file, _ := os.Open(path)
			io.Copy(writer, file)
			file.Close()
		}
		return nil
	})
}

// UnZip 解压
func UnZip(zipFile string, destDir string, encodeing string) (string, error) {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return "", err
	}
	defer zipReader.Close()

	var path string

	for index, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.Flags == 0 {
			name, err := charsetx.Decode([]byte(f.Name), encodeing)
			if err != nil {
				return "", err
			}

			fpath = filepath.Join(destDir, name)
		}

		if index == 0 {
			path = filepath.Dir(fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return "", err
			}
			inFile, err := f.Open()
			if err != nil {
				return "", err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return "", err
			}

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return "", err
			}
			inFile.Close()
			outFile.Close()
		}
	}
	return path, nil
}

// UnZipFile 解压文件，返回文件名和对应本地路径
func UnZipFile(zipFile string, destDir string, encodeing string) (map[string]string, error) {

	fileMap := make(map[string]string)

	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		var fpath = ""
		if f.Flags == 0 {
			name, err := charsetx.Decode([]byte(f.Name), encodeing)
			if err != nil {
				return nil, err
			}
			fpath = filepath.Join(destDir, name)
			fileMap[name] = fpath
		} else {
			fpath = filepath.Join(destDir, f.Name)
			fileMap[f.Name] = fpath
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return nil, err
			}
			inFile, err := f.Open()
			if err != nil {
				return nil, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return nil, err
			}

			inFile.Close()
			outFile.Close()
		}
	}
	return fileMap, nil
}

// ReadFile 读取JSON文件内容到结构体
func ReadFile(path string, result interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err)
		return err
	}

	if len(b) > 0 {
		err = json.Unmarshal(b, result)
		if err != nil {
			fmt.Print(err)
			return err
		}
	} else {
		return errors.New("not found data")
	}

	return nil
}

// Mkdir 调用os.MkdirAll递归创建文件夹
func Mkdir(filePath string) error {
	if !isExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			fmt.Println("创建文件夹失败,error info:", err)
			return err
		}
		return err
	}
	return nil
}

// 判断所给路径文件/文件夹是否存在(返回true是存在)
func isExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
