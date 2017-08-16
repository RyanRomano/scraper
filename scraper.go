package main

import "net/http"
import "io"
import "os"
import "fmt"
import "strconv"
import "strings"
import "bufio"
import "os/exec"
import "path/filepath"


var(
  resp *http.Response
  w    *bufio.Writer
  URL = ""
)
func main() {
  URL = os.Args[1]
  fileOUT, err := os.Create("output.txt")
  if err != nil {
  	fmt.Println(err)
  }
  w = bufio.NewWriter(fileOUT)

	for i := 138; i < 150; i++ {
		s := strconv.Itoa(i)

		fileNameParts := []string{"part-",s,".ts"}
		fileNameString := strings.Join(fileNameParts, "")
		fileNameStringFixedParts := []string{"fixed-", fileNameString}
		fileNameStringFixed := strings.Join(fileNameStringFixedParts, "")

    resp = downloadChunk(s)
    createFile(fileNameString, resp)
    writeOutChunk(fileNameStringFixed)
    fixChunkQuality(fileNameString, fileNameStringFixed)
    deleteOriginal(fileNameString)
	}
  concatChunks()
  deleteParts()
}

func downloadChunk(s string) (*http.Response){
  urlParts := []string{URL,s,"-v1-a1.ts"}

  chunkUrl:= strings.Join(urlParts, "")
  resp, err := http.Get(chunkUrl)
  if err != nil {
    fmt.Println(err)
  }
  return resp
}

func createFile(fileNameString string, resp *http.Response){
  out, err := os.Create(fileNameString)
  if err != nil {
      fmt.Println(err)
  }
  io.Copy(out, resp.Body)
  defer resp.Body.Close()
  defer out.Close()
}

func fixChunkQuality(fileNameString string, fileNameStringFixed string) {
  fixCmdArguments := []string{
    "-i", fileNameString, "-c:v", "libx264",
    "-crf", "20", "-preset", "veryfast" , "-c:a", "libmp3lame",
    "-b:a", "320k", "-af", "apad", "-shortest", "-avoid_negative_ts",
    "make_zero", "-fflags", "+genpts", "-strict", "-2", fileNameStringFixed}
  fixCmd := exec.Command("ffmpeg", fixCmdArguments...)
  err := fixCmd.Run()
  if err != nil {
    fmt.Println(err)
  }
}


func deleteOriginal(fileNameString string){
  deleteCmd := exec.Command("rm", fileNameString)
  err := deleteCmd.Run()
  if err != nil {
    fmt.Println(err)
  }
}

func deleteParts(){
  parts, err := filepath.Glob("fixed*")
  if err != nil {
    fmt.Println(err)
  }
  for _, p := range parts {
    deletePartsCmd := exec.Command("rm", p)
    err := deletePartsCmd.Run()
    if err != nil {
      fmt.Println(err)
    }
  }
  deleteTxtFile := exec.Command("rm", "output.txt")
  err = deleteTxtFile.Run()
  if err != nil {
    fmt.Println(err)
  }
}

func concatChunks(){
  concatCmdArguments := []string{"-f","concat","-i","output.txt","-c","copy","ENTIRE-EPISODE.ts"}
  concatCmd := exec.Command("ffmpeg", concatCmdArguments...)
  err := concatCmd.Run()
  if err != nil {
      fmt.Println(err)
  }
}

func writeOutChunk(fileNameStringFixed string){
  _, err := fmt.Fprintf(w, "file '" + fileNameStringFixed + "'\n")
  if err != nil {
      fmt.Println(err)
  }
  w.Flush()
}
