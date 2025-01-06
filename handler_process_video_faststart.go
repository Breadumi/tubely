package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func processVideoForFastStart(filePath string) (string, error) {

	outputPath := filePath + ".processing"
	fmt.Println(filePath)
	fmt.Println(outputPath)
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	cmd2 := exec.Command("ffprobe", "-v", "trace", "-show_format", "-print_format", "json", "-show_streams", outputPath)

	// Capture standard error
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var stderr2 bytes.Buffer
	cmd2.Stderr = &stderr2
	outputCmd, err := os.Create("output.txt")
	if err != nil {
		fmt.Println(err)
	}
	cmd2.Stdout = outputCmd
	defer outputCmd.Close()

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %v, stderr: %s", err, stderr.String())
	}
	err = cmd2.Run()
	if err != nil {
		return "", fmt.Errorf("ffprobe error: %v, stderr: %s", err, stderr2.String())
	}

	fmt.Println(outputPath)

	return outputPath, nil
}
