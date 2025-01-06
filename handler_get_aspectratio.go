package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	type Result struct {
		Streams []struct {
			Width  int `json:"width,omitempty"`
			Height int `json:"height,omitempty"`
		} `json:"streams"`
	}

	result := Result{}
	decoder := json.NewDecoder(&buffer)
	if err := decoder.Decode(&result); err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("error decoding JSON from ffprobe: %s", err)
	}

	height := float64(result.Streams[0].Height)
	width := float64(result.Streams[0].Width)

	aspectRatio := "other"
	landscapeRatio := 1.777
	portraitRatio := 0.5625

	if math.Abs(width/height-portraitRatio) < 0.1 {
		aspectRatio = "9:16"
	} else if math.Abs(width/height-landscapeRatio) < 0.1 {
		aspectRatio = "16:9"
	}

	return aspectRatio, nil

}
