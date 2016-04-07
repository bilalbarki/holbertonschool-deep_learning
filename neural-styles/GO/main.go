package main

import "net/http"
import "html/template"
import "io"
import (
       "strconv"
       "fmt"
       "path/filepath"
       "math/rand"
       "os"
       "bytes"
       "encoding/base64"
       "io/ioutil"
       "log"
       "net/smtp"
       "os/exec"
)

var uploadTemplate, _ = template.ParseFiles("upload.html")
var errorTemplate, _ = template.ParseFiles("error.html")

func check(err error) { if err != nil { panic(err) } }

func errorHandler(fn http.HandlerFunc) http.HandlerFunc {
     return func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                         if e, ok := recover().(error); ok {
                                  w.WriteHeader(500)
                                                        errorTemplate.Execute(w, e)
                                                                                        }
                                                                                                }()
                                                                                                        fn(w, r)
                                                                                                        }
}

func upload(w http.ResponseWriter, r *http.Request) {
     if r.Method != "POST" {
        uploadTemplate.Execute(w, nil)
                                  return
                                  }
                                  f, h, err := r.FormFile("image")
                                  check(err)
                                  fileext:=filepath.Ext(h.Filename)
                                  defer f.Close()
                                  style:=r.FormValue("image_style")
                                  fmt.Println(style)
                                  email:=r.FormValue("email")
                                  fmt.Println(email)
                                  file_name:="image-"+strconv.Itoa(rand.Intn(1000))+fileext
                                  fmt.Println(file_name)
                                  t, err := os.Create("./"+file_name)

                                  check(err)
                                  defer t.Close()
                                  _, err = io.Copy(t, f)
                                  check(err)
                                  uploadTemplate.Execute(w, nil)
                                  w.Write([]byte("you will get a mail shortly"))
                                  //http.Redirect(w, r, "/view?id="+t.Name()[6:], 302)
                                  //cmd := "th"
                                  //th neural_style.lua -style_image image-81.jpg -content_image 81.jpg -gpu -1 -image_size 20
                                  //args := []string{"neural_style.lua", "?"}
                                 // if err := exec.Command(cmd,args...).Run(); err != nil {
                                     //fmt.Println(err)
                                     //                      os.Exit(1)
                                //                           }
                                //cmd:=exec.Command("th","neural_style.lua -style_image 81.jpg -content_image image-81.jpg -gpu -1 -image_size 20").Start()
                                go func(){
                                cmd := exec.Command("th", "neural_style.lua", "-style_image", style, "-content_image", file_name, "-gpu", "-1", "-image_size",
"256")
                                cmd.Start()
                                cmd.Wait()
                                sendMail("out.png",email)}()
}

/*
func view(w http.ResponseWriter, r *http.Request) {
     w.Header().Set("Content-Type", "image")
     http.ServeFile(w, r, "image-"+r.FormValue("id"))
}*/

func main() {
     http.HandleFunc("/", errorHandler(upload))
     //http.HandleFunc("/view", errorHandler(view))
     http.ListenAndServe(":9000", nil)
}


func sendMail(fname string, email string){
var buf bytes.Buffer
auth := smtp.PlainAuth("", "bilalbarki@hotmail.com", "Totallymadami7", "smtp-mail.outlook.com")
//set necessary variables
from := "bilalbarki@hotmail.com"
to := email
to_name := "first last"
marker := "ACUSTOMANDUNIQUEBOUNDARY"
subject := "Check out my test email"
body := "Your requested file is attached" //for HTML emails just put HTML in the body
file_location := fname
file_name := file_location

//part 1 will be the mail headers
part1 := fmt.Sprintf("From: Example <%s>\r\nTo: %s <%s>\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n--%s", from, to_name, to, subject, marker, marker)

//part 2 will be the body of the email (text or HTML)
part2 := fmt.Sprintf("\r\nContent-Type: text/html\r\nContent-Transfer-Encoding:8bit\r\n\r\n%s\r\n--%s", body, marker)

//read and encode attachment
content, _ := ioutil.ReadFile(file_location)
encoded := base64.StdEncoding.EncodeToString(content)

//split the encoded file in lines (doesn't matter, but low enough not to hit a max limit)
lineMaxLength := 500
nbrLines := len(encoded) / lineMaxLength

//append lines to buffer
for i := 0; i < nbrLines; i++ {
buf.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n")
} //for

//append last line in buffer
buf.WriteString(encoded[nbrLines*lineMaxLength:])

//part 3 will be the attachment
part3 := fmt.Sprintf("\r\nContent-Type: application/csv; name=\"%s\"\r\nContent-Transfer-Encoding:base64\r\nContent-Disposition: attachment; filename=\"%s\"\r\n\r\n%s\r\n--%s--", file_location, file_name, buf.String(), marker)

//send the email
err := smtp.SendMail("smtp-mail.outlook.com:587", auth, from, []string{to}, []byte(part1+part2+part3))

//check for SendMail error
if err != nil {
log.Fatal(err)
} //if
}
