package main

import "github.com/conformal/gotk3/gtk"
import "gitlab.com/seriyyy95/morph"
import "log"
import "os"
import "path/filepath"
import "unicode/utf8"
import "strings"
import "strconv"

type resultData struct {
    resultString string
    countKeywords int
    countResult int
}

func GetPath() string {
    dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
            log.Fatal(err)
    }
    return dir
}

func ReadTextView(object *gtk.TextView) string {
    buffer, err := object.GetBuffer()
    if err != nil {
		    log.Fatal(err)
    }
    start, end := buffer.GetBounds()
    text, err := buffer.GetText(start, end, true)
	if err != nil {
	    log.Fatal("Unable to get text:", err)
	}
    return text;
}

func WriteTextView(object *gtk.TextView, text string){
    buffer, err := object.GetBuffer()
    if err != nil {
		    log.Fatal(err)
    }
    buffer.SetText(text)
}

func GetWindow(builder *gtk.Builder) *gtk.Window {
    object, err := builder.GetObject("MainWindow")
    if err != nil {
		log.Fatal(err)
	}
    window, ok := object.(*gtk.Window);
    if(!ok) {
        log.Fatal("Can`t cast window object");
    }
    return window
}

func GetTextView(builder *gtk.Builder, name string) *gtk.TextView {
    object, err := builder.GetObject(name)
    if err != nil {
		log.Fatal(err)
	}
    element, ok := object.(*gtk.TextView);
    if(!ok) {
        log.Fatal("Can`t cast object to element");
    }
    return element
}

func GetButton(builder *gtk.Builder, name string) *gtk.Button {
    buttonObject, err := builder.GetObject(name)
    if err != nil {
		log.Fatal(err)
	}
    button, ok := buttonObject.(*gtk.Button);
    if(!ok) {
        log.Fatal("Can`t cast button object");
    }
    return button
}

func GetProgressBar(builder *gtk.Builder, name string) *gtk.ProgressBar {
    barObject, err := builder.GetObject(name)
    if err != nil {
		log.Fatal(err)
	}
    bar, ok := barObject.(*gtk.ProgressBar);
    if(!ok) {
        log.Fatal("Can`t cast progress object");
    }
    return bar
}

func GetLabel(builder *gtk.Builder, name string) *gtk.Label {
    object, err := builder.GetObject(name)
    if err != nil {
		log.Fatal(err)
	}
    label, ok := object.(*gtk.Label);
    if(!ok) {
        log.Fatal("Can`t cast label object");
    }
    return label
}


func GetBaseForm(word string) string{
    _, norms, _ := morph.Parse(word)
    if len(norms) == 0{
        return word
    }else{
        return norms[0]
    }
}

func NormalizeKeyword(keyword string) string {
    keywordArray := strings.Split(keyword," ")
    var resultArray []string
    for _,word := range keywordArray {
       resultArray = append(resultArray, GetBaseForm(word))
    }
    return strings.Join(resultArray, " ")
}

func StringInSlice(needle string, haystack []string) bool {
    for _, str := range haystack {
        if str == needle {
            return true
        }
    }
    return false
}

func ProccessData(mainKeyword string, inputKeywords string, result chan resultData, progress chan float64){
    normMainKeyword := NormalizeKeyword(mainKeyword)
    log.Printf("Normalized Keyword: %s", normMainKeyword)
    inputKeywordsArray := strings.Split(inputKeywords, "\n")
    totalKeywords := len(inputKeywordsArray)
    var resultKeywords []string
    var parsent float64
    for index, keyword := range inputKeywordsArray {
        parsent = (float64(index+1) * 100 / float64(totalKeywords)) / 100
        progress <- parsent
        keywordArray := strings.Split(keyword, " ")
        for _, word := range keywordArray {
            if utf8.RuneCountInString(word) <= 3 {
                log.Printf("Word %s too short, skipped", word)
                continue
            }
            if word != normMainKeyword {
                newWord := normMainKeyword + " " + word
                if !StringInSlice(newWord, resultKeywords){
                    resultKeywords = append(resultKeywords, newWord)
                    log.Printf("Word %s added", newWord)
                }else{
                    log.Printf("Word %s already exists, skipped", newWord)
                }
            }
        }
    }
    resultString := strings.Join(resultKeywords, "\n")
    var data resultData
    data.resultString = resultString
    data.countKeywords = totalKeywords
    data.countResult = len(resultKeywords)
    result <- data
}

func PrintData(field *gtk.TextView, totalCountLabel *gtk.Label, resultCountLabel *gtk.Label, result chan resultData) {
    var data resultData
    data = <-result
    WriteTextView(field, data.resultString)
    totalCountLabel.SetText(strconv.Itoa(data.countKeywords))
    resultCountLabel.SetText(strconv.Itoa(data.countResult))
}

func UpdateProgress(bar *gtk.ProgressBar, progress chan float64){
    for {
        var parsent float64
        parsent = <-progress
        bar.SetFraction(parsent)
        if parsent >= 1 {
            break
        }
    }
}

func main(){
    err := morph.Init();
    if err != nil {
        log.Fatal(err)
    }

	gtk.Init(nil)
    builder, err := gtk.BuilderNew()
    if err != nil {
		log.Fatal(err)
	}
    err = builder.AddFromFile(GetPath() + "/interface.glade")
    if err != nil {
		log.Fatal(err)
	}
    window := GetWindow(builder)
    window.Connect("destroy", func() {
		gtk.MainQuit()
	})
    window.SetDefaultSize(800, 800)
    window.SetTitle("Keyword Mixer")
    window.ShowAll()
    mainKeywordField := GetTextView(builder, "MainKeyword")
    inputKeywordsField := GetTextView(builder, "InputKeywords")
    resultKeywordsField := GetTextView(builder, "ResultKeywords")
    proccessButton := GetButton(builder, "ProccessButton")
    progressBar := GetProgressBar(builder, "ProgressBar")
    totalCountLabel := GetLabel(builder, "TotalCount")
    resultCountLabel := GetLabel(builder, "ResultCount")

    var resultChannel chan resultData = make(chan resultData)
    var progressChannel chan float64 = make(chan float64)
    defer close(resultChannel)
    defer close(progressChannel)

    proccessButton.Connect("clicked", func(){
        mainKeyword := ReadTextView(mainKeywordField)
        inputKeywords := ReadTextView(inputKeywordsField)
        go ProccessData(mainKeyword, inputKeywords, resultChannel, progressChannel)
        go PrintData(resultKeywordsField, totalCountLabel, resultCountLabel, resultChannel)
        go UpdateProgress(progressBar, progressChannel)
    })

    gtk.Main()
}
