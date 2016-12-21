package main

import "fmt"
import "html/template"
import "net/http"
import "database/sql"
import "log"
import "time"
import _ "github.com/mattn/go-sqlite3"
import "os"

type post struct {
    id int
    name string
    content string
    timestamp int64
}

type Template_Post struct {
    Name string
    Content string
    Timestamp string
}

func (p *post) Template() Template_Post {
    return Template_Post{p.name, p.content, time.Unix(p.timestamp, 0).Format(time.UnixDate)}
}

func (p *post) String() string {
    return fmt.Sprintf("post(%d, \"%s\", \"%s\", %d)", p.id, p.name, p.content, p.timestamp)
}

func main() {

    db := openDb("./posts.db")
    defer db.Close()
    initializePostsTable(db)

    t, err := template.ParseFiles("templates/posts.html")
    if err != nil {
        fmt.Println(err)
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Handling an http thingy I guess")
        if r.Method == "GET" {
            showPosts(w, r, db, t)
        } else if r.Method == "POST" {
            makePost(w, r, db)
            redirect_home(w, r)
        } else {
            fmt.Fprintf(w, "Invalid HTTP method!\n")
        }
    })
    portstr := ":"
    if len(os.Args) > 1 {
        portstr += os.Args[1]
    } else {
        portstr += "8080"
    }
    err = http.ListenAndServe(portstr, nil)
    fmt.Println(err)

}

func redirect_home(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/", 303)
}

func showPosts(w http.ResponseWriter, r *http.Request, db *sql.DB, t *template.Template) {
    posts := getPosts(db)
    template_posts := make([]Template_Post, 0)
    for _, p := range posts {
        template_posts = append(template_posts, p.Template())
    }
    err := t.Execute(w, template_posts)
    if err != nil {
        fmt.Println(err)
    }
}

func makePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    name := r.FormValue("name")
    content := r.FormValue("content")
    if len(name) > 0 && len(name) <= 15 && len(content) > 0 && len(content) <= 140 {
        addPost(db, &post{name: name, content: content, timestamp: time.Now().Unix()})
    }
}

func openDb(filename string) (*sql.DB) {
    db, err := sql.Open("sqlite3", filename)
    if err != nil {
        log.Fatal(err)
    }
    return db
}

func initializePostsTable(db *sql.DB) {
    sqlStmt := `
    create table if not exists posts (id integer not null primary key, name text, content text, time datetime);
    `
    _, err := db.Exec(sqlStmt)
    if err != nil {
        log.Printf("%q: %s\n", err, sqlStmt)
        return
    }
}

func addPost(db *sql.DB, p *post) {
    tx, err := db.Begin()
    if err != nil {
        log.Fatal(err)
    }
    stmt, err := tx.Prepare("insert into posts (name, content, time) values(?, ?, ?)")
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()
    _, err = stmt.Exec(p.name, p.content, p.timestamp)
    if err != nil {
        log.Fatal(err)
    }
    tx.Commit()
}

func getPosts(db *sql.DB) ([]post){
    rows, err := db.Query("select id, name, content, time from posts order by time desc")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    posts := make([]post, 0)
    for rows.Next() {
        var p post
        var t time.Time
        err = rows.Scan(&p.id, &p.name, &p.content, &t)
        p.timestamp = t.Unix()
        if err != nil {
            log.Fatal(err)
        }
        posts = append(posts, p)
    }
    err = rows.Err()
    if err != nil {
        log.Fatal(err)
    }
    return posts
}
