package main

func processVideoForFastStart(filePath string) (string, error) {

	outputPath := filePath + ".processing"
	cmd.Exec()
	a := cmd.Exec("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)

	return "", nil
}
