package filex

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/charsetx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/utils/storage"

	storagecli "rxcsoft.cn/utils/storage/client"
)

var (
	picTypes = []string{"image/jpeg", "image/png"}
	csvTypes = []string{
		"text/plain",
		"text/plain; charset=utf-8",
		"application/octet-stream",
		"text/csv",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-excel.sheet.macroEnabled.12",
		"application/vnd.ms-excel.template.macroEnabled.12",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.template",
	}
	docTypes = []string{
		"application/msword",
		"application/vnd.ms-word.document.macroEnabled.12",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.ms-excel.sheet.macroEnabled.12",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.ms-powerpoint.presentation.macroEnabled.12",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/pdf",
	}
	zipTypes = []string{"application/zip", "application/x-zip-compressed"}
)

type CopyFile struct {
	Domain       string
	OldApp       string
	NewApp       string
	DatastoreMap map[string]string
}

// SaveFile 创建并保存文件到minio
func SaveFile(data []byte, domain, fileName, fileType, contentType, appRoot string) (file *storage.ObjectInfo, e error) {
	err := Mkdir("temp/")
	if err != nil {
		return nil, err
	}

	// 创建临时文件
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return nil, err
	}

	fo, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fo.Close()

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		return nil, err
	}
	filePath := path.Join(appRoot, fileType, fileName)
	result, err := minioClient.SavePublicObject(fo, filePath, contentType)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SaveLocalFile 创建并保存文件本地
func SaveLocalFile(data []byte, fileName string) (e error) {
	// 创建临时文件
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

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

// DeletePublicHeaderFile 删除头像或LOGO文件
func DeletePublicHeaderFile(domain, fileName string) (d, f string, e error) {
	// 删除对象名编辑
	if strings.Contains(fileName, "/") {
		// 带路径文件名的场合
		index := strings.Index(fileName, domain)
		if index != -1 {
			// 全路径场合
			fileName = fileName[index+len(domain)+1:]
		}
	} else {
		// 单纯文件名的场合
		fileName = path.Join("public", "header", fileName)
	}

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		fmt.Printf("delete public header file has error :%v", err)
		return domain, fileName, err
	}

	// 获取文件对象的信息
	fileInfo, err := minioClient.GetObjectInfo(fileName)
	if err != nil {
		fmt.Printf("delete public header file has error :%v", err)
		return domain, fileName, err
	}
	// 删除文件对象
	if err := minioClient.DeleteObject(fileName); err != nil {
		fmt.Printf("delete public header file has error :%v", err)
		return domain, fileName, err
	}
	// 文件删除成功后,修改顾客的已使用存储空间的大小(开发除外)
	if domain != "system" && domain != "proship.co.jp" {
		err = ModifyUsedSize(domain, -float64(fileInfo.Size))
		if err != nil {
			fmt.Printf("delete public header file has error :%v", err)
			return domain, fileName, err
		}
	}

	return domain, fileName, nil
}

//获取minio服务器上的图片文件信息
func GetMinioHeaderInfo(domain, fileName string) (*storage.ObjectInfo, error) {
	// 删除对象名编辑
	if strings.Contains(fileName, "/") {
		// 带路径文件名的场合
		index := strings.Index(fileName, domain)
		if index != -1 {
			// 全路径场合
			fileName = fileName[index+len(domain)+1:]
		}
	} else {
		// 单纯文件名的场合
		fileName = path.Join("public", "header", fileName)
	}

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		fmt.Printf("get public header fileinfo has error :%v", err)
		return nil, err
	}
	obj, err := minioClient.GetObjectInfo(fileName)
	if err != nil {
		fmt.Printf("get public header fileinfo has error :%v", err)
		return nil, err
	}
	return obj, nil
}

// DeletePublicDataFile 删除文件类型字段数据的文件
func DeletePublicDataFile(domain, appID, fileName string) (d, f string, e error) {
	// 删除对象名编辑
	if strings.Contains(fileName, "/") {
		// 带路径文件名的场合
		index := strings.Index(fileName, domain)
		if index != -1 {
			// 全路径场合
			fileName = fileName[index+len(domain)+1:]
		}
	} else {
		// 单纯文件名的场合
		appRoot := "app_" + appID
		fileName = path.Join("public", appRoot, "data", fileName)
	}

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		fmt.Printf("delete public data file has error :%v", err)
		return domain, fileName, err
	}

	// 获取文件对象的信息
	fileInfo, err := minioClient.GetObjectInfo(fileName)
	if err != nil {
		fmt.Printf("delete public data file has error :%v", err)
		return domain, fileName, err
	}
	// 删除文件对象
	if err := minioClient.DeleteObject(fileName); err != nil {
		fmt.Printf("delete public data file has error :%v", err)
		return domain, fileName, err
	}
	// 文件删除成功后,修改顾客的已使用存储空间的大小
	err = ModifyUsedSize(domain, -float64(fileInfo.Size))
	if err != nil {
		fmt.Printf("delete public data file has error :%v", err)
		return domain, fileName, err
	}

	return domain, fileName, nil
}

// CopyPublicDataFile 拷贝文件类型字段数据文件
func CopyPublicDataFile(domain, appID, fileName string) (d, f string, e error) {
	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		fmt.Printf("copy public data file has error :%v", err)
		return domain, fileName, err
	}

	// 拷贝元对象名编辑
	if strings.Contains(fileName, "/") {
		// 带路径文件名的场合
		index := strings.Index(fileName, domain)
		if index != -1 {
			// 全路径场合
			fileName = fileName[index+len(domain)+1:]
		}
	} else {
		// 单纯文件名的场合
		appRoot := "app_" + appID
		fileName = path.Join("public", appRoot, "data", fileName)
	}

	// 拷贝先对象名编辑
	timestamp := time.Now().Format("20060102150405")
	oldSingleName := filepath.Base(fileName)
	nameIndex := strings.Index(oldSingleName, "_")
	newSingleName := timestamp + oldSingleName[nameIndex:]
	newFileName := strings.ReplaceAll(fileName, oldSingleName, newSingleName)

	// 拷贝文件对象
	result, err := minioClient.CopyObject(fileName, newFileName)
	if err != nil {
		fmt.Printf("copy public data file has error :%v", err)
		return domain, fileName, err
	}

	// 判断顾客上传文件是否在设置的最大存储空间以内
	canUpload := CheckCanUpload(domain, float64(result.Size))
	if canUpload {
		// 如果没有超出最大值，就对顾客的已使用大小进行累加
		err = ModifyUsedSize(domain, float64(result.Size))
		if err != nil {
			fmt.Printf("copy public data file has error :%v", err)
			return domain, fileName, err
		}
	} else {
		// 如果已达上限，则删除刚才上传的文件
		minioClient.DeleteObject(result.Name)
		err = errors.New("最大ストレージ容量に達しました。ファイルのアップロードに失敗しました")
		fmt.Printf("copy public data file has error :%v", err)
		return domain, fileName, err
	}

	return domain, result.MediaLink, nil
}

// DeletePublicDataFiles 删除多个文件类型字段数据的文件
func DeletePublicDataFiles(domain, appID string, fileNameList []string) (d string, files []string, e error) {
	fs := []string{}
	delSize := 0.0
	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		fmt.Printf("delete public data file has error :%v", err)
		return domain, fs, err
	}

	// 循环删除
	for _, fileName := range fileNameList {
		// 删除对象名编辑
		if strings.Contains(fileName, "/") {
			// 带路径文件名的场合
			index := strings.Index(fileName, domain)
			if index != -1 {
				// 全路径场合
				fileName = fileName[index+len(domain)+1:]
			}
		} else {
			// 单纯文件名的场合
			appRoot := "app_" + appID
			fileName = path.Join("public", appRoot, "data", fileName)
		}

		// 获取文件对象的信息
		fileInfo, err := minioClient.GetObjectInfo(fileName)
		if err != nil {
			if len(fs) > 0 {
				// 部分文件删除成功后,修改顾客的已使用存储空间的大小
				err = ModifyUsedSize(domain, -float64(delSize))
				if err != nil {
					fmt.Printf("delete public data file has error :%v", err)
					return domain, fs, err
				}
			}
			fmt.Printf("delete public data file has error :%v", err)
			return domain, fs, err
		}

		if err := minioClient.DeleteObject(fileName); err != nil {
			if len(fs) > 0 {
				// 部分文件删除成功后,修改顾客的已使用存储空间的大小
				err = ModifyUsedSize(domain, -float64(delSize))
				if err != nil {
					fmt.Printf("delete public data file has error :%v", err)
					return domain, fs, err
				}
			}
			fmt.Printf("delete public data file has error :%v", err)
			return domain, fs, err
		}
		// 删除成功文件集合
		fs = append(fs, fileName)
		// 删除成功文件集合大小累计
		delSize += float64(fileInfo.Size)
	}

	// 文件删除成功后,修改顾客的已使用存储空间的大小
	err = ModifyUsedSize(domain, -float64(delSize))
	if err != nil {
		fmt.Printf("delete public data file has error :%v", err)
		return domain, fs, err
	}

	return domain, fs, nil
}

// DeleteDatastoreFiles 删除台账文件夹下的所有文件
func DeleteDatastoreFiles(domain, appID, datastoreID string) (d string, files []string, e error) {
	fs := []string{}
	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		fmt.Printf("delete public data file has error :%v", err)
		return domain, fs, err
	}

	// 编辑文件路径
	appRoot := "app_" + appID
	datastoreUrl := "datastore_" + datastoreID
	filePath := path.Join(appRoot, "data", datastoreUrl)

	delSize, err := minioClient.DeletePath(filePath)
	if err != nil {
		fmt.Printf("delete public data file has error :%v", err)
		return domain, fs, err
	}
	// 文件删除成功后,修改顾客的已使用存储空间的大小
	err = ModifyUsedSize(domain, -float64(delSize))
	if err != nil {
		fmt.Printf("delete public data file has error :%v", err)
		return domain, fs, err
	}

	return domain, fs, nil
}

//删除dev端模板文件
func DeleteMinioTemplateBackups(domain, fileName string) error {

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		fmt.Printf("delete public header file has error :%v", err)
		return err
	}
	if err := minioClient.DeleteObject(fileName); err != nil {
		fmt.Printf("delete public header file has error :%v", err)
		return err
	}
	return nil
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

	filename := "temp/tmp" + "_" + time.Now().Format("20060102150405") + ".txt"
	f, err := SaveFile(bytesBuffer.Bytes(), domain, filename, "text", "text/plain", appRoot)
	if err != nil {
		return nil
	}

	os.Remove(filename)

	return f
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

// 判断字符串数组是否包含某字符串(返回true是包含)
func isContain(target string, str_array []string) bool {
	sort.Strings(str_array)
	index := sort.SearchStrings(str_array, target)
	if index < len(str_array) && str_array[index] == target {
		return true
	}
	return false
}

// 判断是否支持文件类型(返回true是支持)
func CheckSupport(fileType string, contenType string) bool {
	switch fileType {
	case "pic":
		if isContain(contenType, picTypes) {
			return true
		}
	case "csv":
		if isContain(contenType, csvTypes) {
			return true
		}
	case "zip":
		if isContain(contenType, zipTypes) {
			return true
		}
	case "doc":
		if isContain(contenType, docTypes) || isContain(contenType, zipTypes) || isContain(contenType, csvTypes) || isContain(contenType, picTypes) {
			return true
		}
	default:
		return false
	}

	return false
}

// 判断文件大小是否超标(返回true是未超标)
func CheckSize(domain, fileType string, fileSize int64) bool {

	// 如果是超级管理员用户的情况下，不验证大小。
	if domain == "proship.co.jp" {
		return true
	}

	if fileType == "zip" {
		return fileSize < 1024*1024*1024
	}

	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	var req customer.FindCustomerByDomainRequest
	req.Domain = domain
	response, err := customerService.FindCustomerByDomain(context.TODO(), &req)
	if err != nil {
		return false
	}

	// 获取顾客的上传文件大小限制
	uploadFileSize := response.GetCustomer().GetUploadFileSize()
	if uploadFileSize == 0 {
		uploadFileSize = 5 * 1024 * 1024
	} else {
		uploadFileSize = uploadFileSize * 1024 * 1024
	}

	return fileSize < uploadFileSize
}

// WriteAndSaveFile 写入&保存文件
func WriteAndSaveLocalFile(items []string, timestamp, fileName string) (path string, e error) {
	bytesBuffer := &bytes.Buffer{}

	writer := bufio.NewWriter(bytesBuffer)
	for _, item := range items {
		writer.WriteString(item + "\n")
	}

	writer.Flush() // 此时才会将缓冲区数据写入

	dir := "backups/" + timestamp + "/"
	// 创建文件夹
	e1 := Mkdir(dir)
	if e1 != nil {
		return "", e1
	}

	name := dir + fileName + ".json"
	err := SaveLocalFile(bytesBuffer.Bytes(), name)
	if err != nil {
		return "", err
	}

	return name, nil
}

func ZipBackups(dir, fileName string) {
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
			defer file.Close()
			io.Copy(writer, file)
		}
		return nil
	})
}

// 拷贝台账minio文件
func CopyMinioFile(param CopyFile) {
	minioClient, err := storagecli.NewClient(param.Domain)
	if err != nil {
		loggerx.ErrorLog("Copy datastore minio file fail", err.Error())
		return
	}
	// 编辑路径
	oldAppRoot := "app_" + param.OldApp
	newAppRoot := "app_" + param.NewApp
	oldAppPath := path.Join("public", oldAppRoot)
	newAppPath := path.Join("public", newAppRoot)

	// 复制源app的所有文件到新app下
	size, err := minioClient.CopyPath(oldAppPath, newAppPath, true)
	if err != nil {
		loggerx.ErrorLog("Copy datastore minio file fail", err.Error())
		return
	}
	// 判断复制文件是否超出最大使用空间
	// 判断顾客复制文件是否在设置的最大存储空间以内
	canUpload := CheckCanUpload(param.Domain, float64(size))
	if canUpload {
		// 如果没有超出最大值，就对顾客的已使用大小进行累加
		err = ModifyUsedSize(param.Domain, float64(size))
		if err != nil {
			loggerx.ErrorLog("Copy datastore minio file fail", err.Error())
			return
		}
	} else {
		// 如果已达上限，则删除刚才复制的文件
		_, err := minioClient.DeletePath(newAppPath)
		if err != nil {
			loggerx.ErrorLog("Copy datastore minio file fail", err.Error())
			return
		}
		loggerx.DebugLog("Copy datastore minio file fail", "Maximum storage space has been reached")
	}
	for k, v := range param.DatastoreMap {
		// 编辑台账路径
		oldDatastoreRoot := "datastore_" + k
		newDatastoreRoot := "datastore_" + v
		oldDatastorePath := path.Join(newAppPath, "data", oldDatastoreRoot)
		newDatastorePath := path.Join(newAppPath, "data", newDatastoreRoot)
		// 重命名为新app下的文件
		err = minioClient.RenameFolder(oldDatastorePath, newDatastorePath)
		if err != nil {
			loggerx.ErrorLog("Copy datastore minio file fail", err.Error())
			return
		}
	}

	loggerx.DebugLog("Copy datastore minio file success. App path is:", newAppPath)
}
