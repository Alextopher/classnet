package main

import (
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

func main() {
	// Create default room
	room := rooms.NewRoom("")
	log.Println("Created room", room.code)

	// Build the HTTP router
	router := mux.NewRouter()

	// Logging middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.Method, r.URL)
			next.ServeHTTP(w, r)
		})
	})

	templates := template.Must(template.ParseGlob("templates/*.html"))

	// Serve static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", nil)
	})

	// Room control
	router.HandleFunc("/room/new", func(w http.ResponseWriter, r *http.Request) {
		room := rooms.NewRoom("")
		log.Println("Created room", room.code)
		http.Redirect(w, r, "/room/"+room.code, http.StatusFound)
	})

	// Room is given a code from a form and redirects to the room
	router.HandleFunc("/room/join", func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		log.Println("Redirecting to /room/" + code)
		http.Redirect(w, r, "/room/"+code, http.StatusFound)
	})

	router.HandleFunc("/room/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := mux.Vars(r)["code"]
		templates.ExecuteTemplate(w, "room.html", code)
	})

	// Websocket handler
	router.HandleFunc("/room/{code}/ws", WebsocketHandler)

	// Register Handler
	router.HandleFunc("/room/{code}/register", RegisterHandler)

	// Start the HTTP server
	log.Println("Starting HTTP server")
	log.Println("Listening at http://localhost:8080/room/" + room.code)
	log.Fatal(http.ListenAndServe(":8080", router))
}
