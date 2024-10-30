package main

import (
    "bufio"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "os"
    "strings"
    "text/template"
    "time"
	"strconv"
)

const port = ":3002"

type PageData struct {
    AttemptsRemaining int
    FalseLetter       string
    Display           string
    Message           string
    GameOver          bool
    Word              string
}

var (
    normalWords []string
    hardWords   []string
)

func init() {
    normalWords = loadWords("./mot/motnormal.txt")
    hardWords = loadWords("./mot/mothard.txt")
    rand.Seed(time.Now().UnixNano())
}

func loadWords(filename string) []string {
    file, err := os.Open(filename)
    if err != nil {
        log.Fatalf("Failed to open file: %v", err)
    }
    defer file.Close()

    var words []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        words = append(words, strings.TrimSpace(scanner.Text()))
    }

    if err := scanner.Err(); err != nil {
        log.Fatalf("Error reading file: %v", err)
    }

    return words
}

func getRandomWord(difficulty string) string {
    if difficulty == "hard" {
        return hardWords[rand.Intn(len(hardWords))]
    }
    return normalWords[rand.Intn(len(normalWords))]
}

func Home(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "Home", PageData{})
}

func StartGame(w http.ResponseWriter, r *http.Request) {
    difficulty := r.FormValue("difficulty")
    word := getRandomWord(difficulty)
    display := strings.Repeat("_ ", len(word))

    data := PageData{
        AttemptsRemaining: 6,
        FalseLetter:       "",
        Display:           display,
        Word:              word,
    }

    renderTemplate(w, "Game", data)
}

func PlayGame(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    word := r.FormValue("word")
    display := strings.Split(r.FormValue("display"), " ")
    attemptsRemaining, err := strconv.Atoi(r.FormValue("attempts"))
    if err != nil {
        log.Printf("Error parsing attempts: %v", err)
        attemptsRemaining = 6 // Default to 6 if there's an error
    }
    falseLetter := r.FormValue("falseLetter")
    guess := strings.ToLower(r.FormValue("guess"))

    correct := false
    for i, char := range word {
        if string(char) == guess {
            display[i] = guess
            correct = true
        }
    }

    if !correct {
        attemptsRemaining--
        falseLetter += guess + " "
    }

    gameOver := attemptsRemaining == 0 || !strings.Contains(strings.Join(display, ""), "_")
    message := ""
    if gameOver {
        if attemptsRemaining == 0 {
            message = "Fin de partis ! Le mot était: " + word
        } else {
            message = "Bien jouer ! Tu a trouver le mot: " + word
        }
    }

    data := PageData{
        AttemptsRemaining: attemptsRemaining,
        FalseLetter:       falseLetter,
        Display:           strings.Join(display, " "),
        Message:           message,
        GameOver:          gameOver,
        Word:              word,
    }

    renderTemplate(w, "Game", data)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data PageData) {
    t, err := template.ParseFiles("./template/" + tmpl + ".page.tmpl")
    if err != nil {
        log.Printf("Template parsing error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, data)
    if err != nil {
        log.Printf("Template execution error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func main() {
    http.HandleFunc("/", Home)
    http.HandleFunc("/start", StartGame)
    http.HandleFunc("/play", PlayGame)

    fmt.Println("(http://localhost:3002) - server is starting on port", port)
    err := http.ListenAndServe(port, nil)
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}