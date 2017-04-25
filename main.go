package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const tempDockerFile = "Dockerfile.tmp"

var (
	contextDirectory = "."

	dockerfilePath string
	imageTag       string
)

func init() {
	flag.StringVar(&dockerfilePath, "f", "Dockerfile.dft", "Name of the Dockerfile")
	flag.StringVar(&imageTag, "t", "", "Name and optionally a tag in the 'name:tag' format")
}

type dockerfile struct {
	tag       string
	imageName string
	pullImage bool
	text      []byte
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		contextDirectory = args[0]
	}

	file, err := os.Open(dockerfilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	df, err := parseDockerFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if df.tag == "" {
		fmt.Println("Docker image tag required")
		os.Exit(1)
	}

	buildImage(df)
}

func parseDockerFile(file *os.File) (*dockerfile, error) {
	df := &dockerfile{}
	var textBuffer bytes.Buffer
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			df.text = append(df.text, '\n')
			continue
		}

		if len(line) > 2 && bytes.Equal(line[:2], []byte("//")) {
			textBuffer.WriteByte('#')
			textBuffer.Write(line[2:])
			textBuffer.WriteByte('\n')
			continue
		}

		command := bytes.SplitN(line, []byte{' '}, 2)
		cmd := string(command[0])

		switch cmd {
		case "FROM":
			fromCmd := bytes.SplitN(line, []byte{' '}, 3)
			if len(fromCmd) == 2 {
				df.imageName = string(fromCmd[1])
			} else if len(fromCmd) == 3 {
				df.imageName = string(fromCmd[1])
				df.pullImage = bytes.Equal(fromCmd[2], []byte("pull-always"))
			} else {
				return nil, errors.New("FROM must have an image name and optional pull method")
			}

			textBuffer.Write(fromCmd[0])
			textBuffer.WriteByte(' ')
			textBuffer.Write(fromCmd[1])
			textBuffer.WriteByte('\n')
		case "RUN":
			if len(command) != 2 {
				return nil, errors.New("Invalid RUN command")
			}

			if bytes.Equal(command[1], []byte{'('}) {
				// Special multi-line RUN directive
				runCmd, err := getRunText(scanner)
				if err != nil {
					return nil, err
				}
				textBuffer.Write(command[0])
				textBuffer.WriteByte(' ')
				textBuffer.Write(runCmd)
				textBuffer.WriteByte('\n')
			} else {
				// Normal RUN directive
				textBuffer.Write(line)
				textBuffer.WriteByte('\n')
			}
		case "TAG":
			if len(command) != 2 {
				return nil, errors.New("TAG must have an repository name and optional tag")
			}

			if df.tag != "" {
				return nil, errors.New("TAG used multiple times")
			}

			df.tag = string(command[1])
		default:
			textBuffer.Write(line)
			textBuffer.WriteByte('\n')
		}
	}

	df.text = textBuffer.Bytes()

	if imageTag != "" {
		df.tag = imageTag
	}

	return df, nil
}

func getRunText(scanner *bufio.Scanner) ([]byte, error) {
	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		if bytes.Equal(line, []byte{')'}) {
			break
		}

		if line[len(line)-1] == '\\' {
			buf.Write(line)
			buf.WriteByte('\n')
			continue
		}

		buf.Write(line)
		buf.Write([]byte(" && \\\n"))
	}

	return buf.Bytes()[:buf.Len()-6], nil
}

func buildImage(df *dockerfile) {
	if df.pullImage {
		pullCmd := exec.Command("docker", "pull", df.imageName)
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		if err := pullCmd.Run(); err != nil {
			return
		}
	}

	dockFile := filepath.Join(contextDirectory, tempDockerFile)
	if err := ioutil.WriteFile(dockFile, df.text, 0655); err != nil {
		fmt.Println(err)
		return
	}

	buildCmd := exec.Command("docker", "build", "-t", df.tag, "-f", dockFile, contextDirectory)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Run()

	os.Remove(dockFile)
}
