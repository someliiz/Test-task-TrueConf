package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

const store = `users.json`

var (
	UserNotFound = errors.New("user_not_found")
)

var app *App

type App struct {
	router     chi.Router
	userStore  UserStore
	storePath  string
	routerPath string
	log        *logrus.Entry
}

type AppError struct {
	Err     error  `json:"-"`
	Status  int    `json:"-"`
	Message string `json:"message"`
}

type (
	User struct {
		CreatedAt   time.Time `json:"created_at"`
		DisplayName string    `json:"display_name"`
		Email       string    `json:"email"`
	}
	UserList  map[string]User
	UserStore struct {
		Increment int      `json:"increment"`
		List      UserList `json:"list"`
	}
)

type CreateUserRequest struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type ErrResponse struct {
	Err            error `json:"-"`
	HTTPStatusCode int   `json:"-"`

	StatusText string `json:"status"`
	AppCode    int64  `json:"code,omitempty"`
	ErrorText  string `json:"error,omitempty"`
}

func NewApp(storePath, routerPath string) *App {
	app := &App{
		storePath:  storePath,
		routerPath: routerPath,
	}
	// Создаем экземпляр логгера
	log := logrus.New()
	log.Formatter = &logrus.JSONFormatter{}
	log.Level = logrus.InfoLevel

	// Попытка открыть файл логов
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
		defer file.Close() // Закрыть файл после использования
	} else {
		log.WithError(err).Error("Failed to log to file, using default stderr")
	}

	app.log = logrus.NewEntry(log) // Добавляем логгер в структуру App

	app.initRouter()
	app.initStore()

	return app
}

func handleError(w http.ResponseWriter, r *http.Request, err error, status int, message string) {
	appErr := AppError{Err: err, Status: status, Message: message}
	app.log.Errorf("Error: %v", err)
	render.Status(r, appErr.Status)
	render.JSON(w, r, appErr)
}

func (app *App) initRouter() {
	app.router = chi.NewRouter()
	app.router.Use(middleware.RequestID)
	app.router.Use(middleware.RealIP)
	app.router.Use(middleware.Logger)
	// app.router.Use(middleware.Recoverer)
	app.router.Use(middleware.Timeout(60 * time.Second))

	app.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(time.Now().String()))
	})

	app.router.Route(app.routerPath, app.initUserRoutes)
}

func (app *App) initStore() {
	f, err := ioutil.ReadFile(app.storePath)
	if err != nil {
		// Обработка ошибки чтения файла
		app.log.Error("Failed to read data from file", err)
	}

	err = json.Unmarshal(f, &app.userStore)
	if err != nil {
		// Обработка ошибки разбора JSON
		app.log.Fatal(err)
	}
}
func (app *App) initUserRoutes(r chi.Router) {
	r.Get("/", app.searchUsers)
	r.Post("/", app.createUser)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", app.getUser)
		r.Patch("/", app.updateUser)
		r.Delete("/", app.deleteUser)
	})
}
func (app *App) searchUsers(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, app.userStore.List)
}

func (app *App) createUser(w http.ResponseWriter, r *http.Request) {

}

func (app *App) getUser(w http.ResponseWriter, r *http.Request) {

}

func (app *App) updateUser(w http.ResponseWriter, r *http.Request) {

}

func (app *App) deleteUser(w http.ResponseWriter, r *http.Request) {

}

func (c *CreateUserRequest) Bind(r *http.Request) error { return nil }

func createUser(w http.ResponseWriter, r *http.Request) {
	f, _ := ioutil.ReadFile(store)
	s := UserStore{}
	_ = json.Unmarshal(f, &s)

	request := CreateUserRequest{}

	if err := render.Bind(r, &request); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	s.Increment++
	u := User{
		CreatedAt:   time.Now(),
		DisplayName: request.DisplayName,
		Email:       request.DisplayName,
	}

	id := strconv.Itoa(s.Increment)
	s.List[id] = u

	b, _ := json.Marshal(&s)
	_ = ioutil.WriteFile(store, b, fs.ModePerm)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"user_id": id,
	})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	f, _ := ioutil.ReadFile(store)
	s := UserStore{}
	_ = json.Unmarshal(f, &s)

	id := chi.URLParam(r, "id")

	render.JSON(w, r, s.List[id])
}

type UpdateUserRequest struct {
	DisplayName string `json:"display_name"`
}

func (c *UpdateUserRequest) Bind(r *http.Request) error { return nil }

func updateUser(w http.ResponseWriter, r *http.Request) {
	f, _ := ioutil.ReadFile(store)
	s := UserStore{}
	_ = json.Unmarshal(f, &s)

	request := UpdateUserRequest{}

	if err := render.Bind(r, &request); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	id := chi.URLParam(r, "id")

	if _, ok := s.List[id]; !ok {
		_ = render.Render(w, r, ErrInvalidRequest(UserNotFound))
		return
	}

	u := s.List[id]
	u.DisplayName = request.DisplayName
	s.List[id] = u

	b, _ := json.Marshal(&s)
	_ = ioutil.WriteFile(store, b, fs.ModePerm)

	render.Status(r, http.StatusNoContent)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	f, _ := ioutil.ReadFile(store)
	s := UserStore{}
	_ = json.Unmarshal(f, &s)

	id := chi.URLParam(r, "id")

	if _, ok := s.List[id]; !ok {
		_ = render.Render(w, r, ErrInvalidRequest(UserNotFound))
		return
	}

	delete(s.List, id)

	b, _ := json.Marshal(&s)
	_ = ioutil.WriteFile(store, b, fs.ModePerm)

	render.Status(r, http.StatusNoContent)
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func searchUsers(w http.ResponseWriter, r *http.Request) {
	f, _ := ioutil.ReadFile(store)
	s := UserStore{}
	_ = json.Unmarshal(f, &s)

	render.JSON(w, r, s.List)
}

func main() {
	app := NewApp("users.json", "/api/v1/users")
	http.ListenAndServe(":3333", app.router)
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	//r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(time.Now().String()))
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/users", func(r chi.Router) {
				r.Get("/", searchUsers)
				r.Post("/", createUser)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", getUser)
					r.Patch("/", updateUser)
					r.Delete("/", deleteUser)
				})
			})
		})
	})

	http.ListenAndServe(":3333", r)
}
