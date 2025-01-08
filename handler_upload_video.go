package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)
	videoID, err := uuid.Parse(r.PathValue("videoID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	dbVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find video", err)
		return
	}
	if userID != dbVideo.UserID {
		respondWithError(w, http.StatusUnauthorized, "Invalid User ID", err)
		return
	}

	file, fileHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't form file", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(fileHeader.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed Content-Type header", err)
		return
	}

	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Content-Type must be video/mp4", err)
		return
	}

	tmpFile, err := os.CreateTemp("", "tubely-upload.mp4") // create initial file for processing, overwritten later when processed
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create temporary file", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	_, err = io.Copy(tmpFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could copy file", err)
		return
	}

	tmpFile.Seek(0, io.SeekStart)

	b := make([]byte, 32)
	rand.Read(b)
	bucket := cfg.s3Bucket
	key := hex.EncodeToString(b)

	// create prefix for key
	tmpFilePath := tmpFile.Name()
	aspectRatio, err := getVideoAspectRatio(tmpFilePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve aspect ratio", err)
		return
	}
	switch aspectRatio {
	case "16:9":
		key = "landscape/" + key
	case "9:16":
		key = "portrait/" + key
	case "other":
		key = "other/" + key
	default:
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve aspect ratio err2", err)
		return
	}

	fastStartFilePath, err := processVideoForFastStart(tmpFilePath)
	fmt.Println("This is the the fast start error", err)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't process video to fast start", err)
		return
	}
	tmpFileProcessed, err := os.Open(fastStartFilePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open fast start processed video", err)
		return
	}
	defer os.Remove(fastStartFilePath)
	defer tmpFileProcessed.Close()

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        tmpFileProcessed,
		ContentType: aws.String(mediaType),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't upload to AWS", err)
		return
	}

	videoURL := cfg.s3CfDistribution + "/" + key
	dbVideo.VideoURL = &videoURL

	err = cfg.db.UpdateVideo(dbVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video metadata", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVideo)

}
