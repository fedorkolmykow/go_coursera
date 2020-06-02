package main

import (
	"os"
	"fmt"
	"sort"
	"io"
)

const (
	leaf 		= "├───"
	branch  	= "│	"
	lastleaf 	= "└───"
	empty		= "	"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool)(err error){
	openDir(out, path, "", printFiles)
	return
}

func openDir(out io.Writer, path, shift string, printFiles bool)(err error){
	branc := branch
	lea := leaf

	def_dir, err := os.Open(path)
	defer def_dir.Close()
	if err != nil {
		return err
	}

	files, err:=def_dir.Readdir(0)
	if err != nil {
		return err
	}

	//удаление файлов, если нет параметра -f
	if !printFiles{
		for i := len(files) - 1; i >= 0; i-- {
			if !files[i].IsDir(){
				files = append(files[:i], files[i+1:]...)
			}
		}
	}

	//сортировка
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	for i, val:= range files{
		//если последний элемент, то закрыть ветку
		if (i == len(files)-1) {
			branc = empty
			lea = lastleaf
		}

		if val.IsDir() {
			fmt.Fprintf(out, "%v%v\n", shift+lea, val.Name())
			var nextPath = path + "/" + val.Name()
			openDir(out, nextPath,shift + branc, printFiles)
		} else if (printFiles) {
			if val.Size() != 0 {
				fmt.Fprintf(out, "%v%v (%db)\n", shift+lea, val.Name(), val.Size())
			} else {
				fmt.Fprintf(out, "%v%v (empty)\n", shift+lea, val.Name())
			}
		}
	}
	return
}