package public

import (
	"archive/zip"
	"fmt"
	"github.com/joexu01/container-dispatcher/log"
	"golang.org/x/crypto/bcrypt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func GeneratePwdHash(pwd []byte) (string, error) {
	pwdHash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(pwdHash), nil
}

func ComparePwdAndHash(pwd []byte, hashedPwd string) bool {
	byteHash := []byte(hashedPwd)

	err := bcrypt.CompareHashAndPassword(byteHash, pwd)
	if err != nil {
		log.Error("Comparing hashed password: %s", err.Error())
		return false
	}
	return true
}

//PathExists 判断一个文件或文件夹是否存在
//输入文件路径，根据返回的bool值来判断文件或文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetPathFiles(path string) []string {
	var out []string
	getFiles(path, "", &out)
	return out
}

func getFiles(folder string, prefix string, list *[]string) {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if file.IsDir() {
			getFiles(folder+"/"+file.Name(), prefix+file.Name()+"/", list)
		} else {
			fmt.Println(prefix + file.Name())
			*list = append(*list, prefix+file.Name())
		}
	}
}

// CopyFile copies file from src directory
// to dst directory
func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

type FileTree struct {
	Label    string      `json:"label"`
	Filepath string      `json:"filepath"`
	Type     string      `json:"type"`
	Children []*FileTree `json:"children,omitempty"`
}

func GetFiles2(folder string, prefix string, parent *FileTree) {
	files, _ := ioutil.ReadDir(folder)

	if len(files) == 0 {
		parent.Children = nil
	} else {
		parent.Children = []*FileTree{}
	}

	for _, file := range files {

		if file.IsDir() {
			subNode := &FileTree{
				Label:    file.Name(),
				Filepath: "",
				Children: nil,
			}

			parent.Children = append(parent.Children, subNode)

			GetFiles2(folder+"/"+file.Name(), prefix+file.Name()+"/", subNode)
		} else {
			fmt.Println(prefix + file.Name())
			parent.Children = append(parent.Children, &FileTree{
				Label:    file.Name(),
				Filepath: prefix + file.Name(),
				Children: nil,
			})
		}
	}
}

// GetFilesWithDirInfo 使用此函数注意：第一个传入的 parent 节点的 Label 必须为字符串 "root"
func GetFilesWithDirInfo(folder string, prefix string, parent *FileTree) {
	files, _ := ioutil.ReadDir(folder)

	if len(files) == 0 {
		parent.Children = nil
	} else {
		if parent.Label != "root" {
			parent.Children = []*FileTree{&FileTree{
				Label:    "打包并下载此目录所有文件",
				Filepath: folder,
				Type:     FileTreeNodeTypeDir,
				Children: nil,
			}}
		} else {
			parent.Children = []*FileTree{}
		}
	}

	for _, file := range files {

		if file.IsDir() {
			subNode := &FileTree{
				Label:    file.Name(),
				Filepath: "",
				Children: nil,
			}

			parent.Children = append(parent.Children, subNode)

			GetFilesWithDirInfo(folder+"/"+file.Name(), prefix+file.Name()+"/", subNode)
		} else {
			fmt.Println(prefix + file.Name())
			parent.Children = append(parent.Children, &FileTree{
				Label:    file.Name(),
				Filepath: prefix + file.Name(),
				Children: nil,
				Type:     FileTreeNodeTypeFile,
			})
		}
	}
}

func DirSizeSum(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

//getFileSize get file size by path(B)
func getFileSize(path string) int64 {
	if !exists(path) {
		return 0
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

//exists Whether the path exists
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func Zip(dest string, paths ...string) error {
	zFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zFile.Close()
	zipWriter := zip.NewWriter(zFile)
	defer zipWriter.Close()
	for _, src := range paths {
		// remove the trailing path separator if it is a directory
		srcDir := strings.TrimSuffix(src, string(os.PathSeparator))
		err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// create local file header
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			// set compression method to deflate
			header.Method = zip.Deflate
			// set relative path of file in zip archive
			header.Name, err = filepath.Rel(filepath.Dir(srcDir), path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				header.Name += string(os.PathSeparator)
			}
			// create writer for writing header
			headerWriter, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(headerWriter, f)
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}
