package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	_ "github.com/godror/godror"
)

var db *sqlx.DB
var view View

func Init() {
	var err error

	view = NewView()
	db, err = sqlx.Connect("godror", "")
	if err != nil {
		log.Fatal(err)
	}
}

type Problem struct {
	ID          int    `db:"ID"`
	Title       string `db:"TITLE"`
	Description string `db:"DESCRIPTION"`
}

type ProblemsData struct {
	Problems []Problem
}

type Submission struct {
	ID             int    `db:"ID"`
	ProblemID      int    `db:"PROBLEM_ID"`
	ProblemTitle   string `db:"PROBLEM_TITLE"`
	Solution       string `db:"SOLUTION"`
	StatusTitle    string `db:"STATUS_TITLE"`
	CheckerMessage string `db:"CHECKER_MESSAGE"`
}

type SubmissionsData struct {
	Submissions []Submission
}

type View struct {
	ProblemsT    *template.Template
	SubmissionT  *template.Template
	SubmissionsT *template.Template
	SubmitT      *template.Template
}

func NewView() View {
	var view View
	var err error

	view.ProblemsT, err = template.ParseFiles("layout.html", "problems.html")
	if err != nil {
		log.Fatal(err)
	}

	view.SubmissionT, err = template.ParseFiles("layout.html", "submission.html")
	if err != nil {
		log.Fatal(err)
	}

	view.SubmissionsT, err = template.ParseFiles("layout.html", "submissions.html")
	if err != nil {
		log.Fatal(err)
	}

	view.SubmitT, err = template.ParseFiles("layout.html", "submit.html")
	if err != nil {
		log.Fatal(err)
	}

	return view
}

func ProblemsHandler(w http.ResponseWriter, r *http.Request) {
	var problems []Problem

	const query = "SELECT ID, TITLE, DESCRIPTION FROM PROBLEM ORDER BY ID ASC"
	err := db.Select(&problems, query)
	if err != nil {
		ServerError(w, err)
		return
	}

	err = view.ProblemsT.ExecuteTemplate(w, "layout", ProblemsData{Problems: problems})
	if err != nil {
		ServerError(w, err)
		return
	}
}

func SubmissionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		ServerError(w, err)
		return
	}

	var submission Submission

	const query = "SELECT S.ID, P.TITLE AS PROBLEM_TITLE, S.SOLUTION FROM SUBMISSION S JOIN PROBLEM P ON S.PROBLEM_ID = P.ID WHERE S.ID = :1"
	err = db.Get(&submission, query, id)
	if err != nil {
		ServerError(w, err)
		return
	}

	err = view.SubmissionT.ExecuteTemplate(w, "layout", submission)
	if err != nil {
		ServerError(w, err)
		return
	}
}

func SubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	var submissions []Submission

	const query = "SELECT S.ID, P.ID AS PROBLEM_ID, P.TITLE AS PROBLEM_TITLE, S.SOLUTION, SS.TITLE AS STATUS_TITLE, S.CHECKER_MESSAGE FROM SUBMISSION S JOIN SUBMISSION_STATUS SS ON S.STATUS_ID = SS.ID JOIN PROBLEM P ON S.PROBLEM_ID = P.ID ORDER BY S.ID ASC"
	err := db.Select(&submissions, query)
	if err != nil {
		ServerError(w, err)
		return
	}

	for i := range submissions {
		submissions[i].StatusTitle = statusDescription[submissions[i].StatusTitle]
	}

	err = view.SubmissionsT.ExecuteTemplate(w, "layout", SubmissionsData{Submissions: submissions})
	if err != nil {
		ServerError(w, err)
		return
	}
}

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			ServerError(w, err)
			return
		}

		var problem Problem
		err = db.Get(&problem, "SELECT ID, TITLE, DESCRIPTION FROM PROBLEM WHERE ID = :1", id)
		if err != nil {
			ServerError(w, err)
			return
		}

		err = view.SubmitT.ExecuteTemplate(w, "layout", problem)
		if err != nil {
			ServerError(w, err)
			return
		}
	case "POST":
		problemID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			ServerError(w, err)
			return
		}

		err = r.ParseForm()
		if err != nil {
			ServerError(w, err)
			return
		}

		var submissionID int
		err = db.Get(&submissionID, "SELECT MAX(ID) + 1 FROM SUBMISSION")
		if err != nil {
			ServerError(w, err)
			return
		}

		const query = "INSERT INTO SUBMISSION (ID, USER_ACCOUNT_ID, PROBLEM_ID, SOLUTION) VALUES (:1, 1, :2, :3)"
		_, err = db.Exec(query, submissionID, problemID, r.PostForm.Get("solution"))
		if err != nil {
			ServerError(w, err)
			return
		}

		http.Redirect(w, r, "/submissions", http.StatusSeeOther)
	}
}

func ServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(fmt.Sprintf("Server encountered an error: %s", err)))
	log.Println(err)
}

func main() {
	Init()

	r := mux.NewRouter()

	r.HandleFunc("/problems", ProblemsHandler)
	r.HandleFunc("/problems/{id}", SubmitHandler)
	r.HandleFunc("/submission/{id}", SubmissionHandler)
	r.HandleFunc("/submissions", SubmissionsHandler)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
