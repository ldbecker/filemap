package main

import (
	"bytes"

	"encoding/json"
	"fmt"
	"golang.org/x/exp/slices"
"github.com/codingsince1985/checksum"
	"os"
	"path"
	"strings"
	"time"
)

var typesFound []string

type FileMap map[string]FileInfo

type FileList struct {
	DateMade    int64  `json:"date-made"`
	DirPath     string     `json:"dir-path"`
	FileTypes   []string   `json:"file-types"`
	Files       []FileInfo `json:"files"`
}

type FileInfo struct {
	MD5Sum    string         `json:"md5"`
	Instances []FileInstance `json:"instances"`
}

type FileInstance struct {
	FileType    string `json:"type"`
	FilePath    string `json:"path"`
	FileSize    int64  `json:"size"`
	FileModTime int64  `json:"modified"`
	MD5Sum      string `json:"md5"`
}

func NewFileList(filepath string, filetypes []string) *FileList {
	xtime := time.Now()
	xfiles := make([]FileInfo, 0)
	xList := &FileList{DateMade: xtime.Unix(), Files: xfiles, FileTypes: filetypes, DirPath: filepath}
	return xList

}

func ListFiles(dir string, typelist []string) ([]FileInstance, error) {

	xList := make([]FileInstance, 0)
	files, err := os.ReadDir(dir)
	if err != nil {
		return xList, err
	}
	for _, xfile := range files {
		newfile := FileInstance{}
		xfn := xfile.Name()
		xpath := path.Clean(path.Join(dir, xfn))
		if strings.HasPrefix(xfn, ".") {
			continue
		}
		if xfile.IsDir() {
			dirC, derr := ListFiles(xpath, typelist)
			if derr != nil {
				return xList, derr
			}
			xList = append(xList, dirC...)
			continue
		}
		xfnchunks := strings.Split(xfn, ".")
		xtype := xfnchunks[len(xfnchunks)-1]

		if !slices.Contains(typelist, xtype) && !slices.Contains(typelist, "all") {
			continue
		}
		if !slices.Contains(typesFound, xtype) {
			typesFound = append(typesFound, xtype)
		}
		newfile.FileType = xtype
		newfile.FilePath = xpath
		xstat, xerr := os.Stat(xpath)
		if xerr != nil {
			return xList, xerr
		}

		newfile.FileSize = xstat.Size()
		newfile.FileModTime = xstat.ModTime().Unix()
		/*xbytes, berr := os.ReadFile(xpath)
		if berr != nil {
			return xList, berr
		}
		newfile.MD5Sum = fmt.Sprintf("%x", md5.Sum(xbytes))
		*/
		newsum, serr  := checksum.SHA1sum(xpath)
		if serr != nil {
			return xList, serr
		}
		newfile.MD5Sum = newsum
		xList = append(xList, newfile)

	}
	return xList, nil
}

func main() {
	argz := os.Args
	fmt.Printf("%v\n", os.Args)
	dir := ""
	//types := ""
	typesList := make([]string, 0)
	savepath := ""
	typesFound = make([]string, 0)
	for _, arg := range argz {
		if strings.HasSuffix(arg, "main") {
			continue
		}
		switch aa := strings.Split(arg, "="); aa[0] {
		case arg:
			continue
		case "types":
			typesList = append(typesList, strings.Split(aa[1], ",")...)
		case "dir":
			dir = path.Clean(aa[1])
		case "savepath":

			savepath = path.Clean(path.Join(aa[1], fmt.Sprintf("filelist-%v.json", time.Now().Unix())))
		default:
			fmt.Println("def")
		}
	}
	fmap := make(FileMap)
	fList, err := ListFiles(dir, typesList)
	if err != nil {
		panic(err)
	}

	fmt.Printf("dir=%v\ntypes=%v\nsavepath=%v\ntypesfound=%v\n", dir, typesList, savepath, typesFound)
	//fmt.Printf("%v\n", fList)

	for _, curFile := range fList {
		if xfileInfo, ok := fmap[curFile.MD5Sum]; !ok {
			//fmap[curFile.MD5Sum] = FileInfo{}
			newfinfo := FileInfo{MD5Sum: curFile.MD5Sum, Instances: []FileInstance{curFile}}

			fmap[curFile.MD5Sum] = newfinfo

		} else {
			newfinfo := append(xfileInfo.Instances, curFile)
			newinfo := FileInfo{
				MD5Sum:    curFile.MD5Sum,
				Instances: newfinfo,
			}
			fmap[curFile.MD5Sum] = newinfo
		}

	}
	for _, finfo := range fmap {
		fmt.Printf("%v\n", finfo)
	}
	jsonb, err := json.Marshal(fmap)
	if err != nil {
		panic(err)
	}
	var pbuf bytes.Buffer
	json.Indent(&pbuf, jsonb, "", "  ")
	err = os.WriteFile(savepath, pbuf.Bytes(), 0755)
	if err != nil {
		panic(err)
	}

	escapeds := strings.ReplaceAll(string(jsonb), "\"", "\\\"")
	err = os.WriteFile(savepath + ".txt", []byte(escapeds), 0755)
	if err != nil {
		panic(err)
	}
	finfolist := NewFileList(dir, typesList)
	finfolist.FileTypes =typesFound
	finfolist.DirPath = dir
	finfolist.DateMade = time.Now().Unix()
	finfolist.Files = make([]FileInfo, 0)
	for _, xfinfo := range fmap {
		finfolist.Files = append(finfolist.Files, xfinfo)
	}
	var xbuf bytes.Buffer
	jsonx, err := json.Marshal(finfolist)
	json.Indent(&xbuf, jsonx, "", "  ")
	os.WriteFile(savepath + ".list.json", xbuf.Bytes(), 0755)
}
