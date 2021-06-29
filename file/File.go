package file

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
)

type File struct {
	Name string
}

func (f File) Delete() bool {
	err := os.Remove(f.Name)
	if err != nil {
		return false
	}

	return true
}

func (f File) Create() bool {
	return createFile(f.Name)
}

func (f File) IsExist() bool {
	return existsFile(f.Name)
}

func (f *File) Clear() (*File, error) {
	file, err := os.OpenFile(f.Name, os.O_WRONLY, os.ModePerm)

	if err != nil {
		return f, err
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	err = file.Truncate(0)
	_, err = file.WriteString("")
	if err != nil {
		return nil, err
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	return f, err
}

func (f *File) Write(content string) (*File, error) {
	return f, writeFile(f.Name, content)
}

func (f *File) Append(content string) (*File, error) {
	return f, appendFile(f.Name, content)
}

func (f *File) WriteNewLine(content string) (*File, error) {
	return f, writeNlFile(f.Name, content)
}

func (f *File) AppendNewLine(content string) (*File, error) {
	return f, appendNLFile(f.Name, content)
}

func writeNlFile(name string, str string) error {
	return writeFile(name, str+"\n")
}

func (f File) Read() ([]byte, error) {
	err := createIfNotExist(f.Name)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(f.Name)
}

func (f File) ReadByLine(function func(str string)) error {
	err := createIfNotExist(f.Name)
	if err != nil {
		return err
	}

	file, _ := os.Open(f.Name)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		function(scanner.Text())
	}
	err = file.Close()
	if err != nil {
		return err
	}

	return scanner.Err()
}

func (f File) ReadAllLines() ([]string, error) {
	return readLines(f.Name)
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()
	return lines, scanner.Err()
}

func writeFile(name string, str string) error {
	err := createIfNotExist(name)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(name, os.O_WRONLY, os.ModePerm)

	if err != nil {
		return err
	}

	defer file.Close()
	file.Seek(0, io.SeekStart)
	file.Truncate(0)
	_, err = file.WriteString(str)

	file.Close()
	if err != nil {
		return err
	}

	return nil
}

func appendNLFile(name string, str string) error {
	return appendFile(name, str+"\n")
}

func appendFile(name string, str string) error {
	err := createIfNotExist(name)

	if err != nil {
		return err
	}

	file, err := os.OpenFile(name, os.O_APPEND, os.ModeAppend)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write([]byte(str))
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func createIfNotExist(name string) error {
	if !existsFile(name) {
		if !createFile(name) {
			return errors.New("File not created.")
		}
	}

	return nil
}

func createFile(fileName string) bool {
	file, err := os.Create(fileName)
	if err != nil {
		return false
	}

	err = file.Close()
	if err != nil {
		return false
	}

	return true
}

func existsFile(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}
