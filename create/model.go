package create

const modelDockerfile = `ARG GO_VERSION=1.21.3

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk update && apk add --no-cache ca-certificates openssl git tzdata
RUN update-ca-certificates

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN gofmt -w . && go build -v -o /$1 ./cmd/$1

RUN chmod +x /$1

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/$1 ./$1

ENTRYPOINT ["./$1"]
`

const modelMain = `package main

import (
	"os"
	"os/signal"

	serv "$1/internal/service/$2"
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/envar"
)

func main() {
	envar.SetInt("port", 3000, "Port server", "PORT")
	envar.SetInt("rpc", 4200, "Port rpc server", "RPC_PORT")
	envar.SetStr("dbhost", "localhost", "Database host", "DB_HOST")
	envar.SetInt("dbport", 5432, "Database port", "DB_PORT")
	envar.SetStr("dbname", "", "Database name", "DB_NAME")
	envar.SetStr("dbuser", "", "Database user", "DB_USER")
	envar.SetStr("dbpass", "", "Database password", "DB_PASSWORD")

	serv, err := serv.New()
	if err != nil {
		console.Fatal(err)
	}

	go serv.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	serv.Close()
}
`

const modelService = `package module

import (
	"net/http"

	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/envar"
	"github.com/cgalvisleon/elvis/middleware"
	"github.com/cgalvisleon/elvis/response"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	v1 "$1/internal/service/$2/v1"
)

type Server struct {
	http *http.Server
}

func New() (*Server, error) {	
	server := Server{}

	port := envar.GetInt(3300, "PORT")
	if port != 0 {
		r := chi.NewRouter()

		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		latest := v1.New()

		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			response.HTTPError(w, r, http.StatusNotFound, "404 Not Found")
		})

		r.Mount("/", latest)
		r.Mount("/v1", latest)

		handler := cors.AllowAll().Handler(r)
		addr := strs.Format(":%d", port)
		serv := &http.Server{
			Addr:    addr,
			Handler: handler,
		}

		server.http = serv
	}
	
	return &server, nil
}

func (serv *Server) Close() {
	v1.Close()

	console.LogK("Http", "Shutting down server...")
}

func (serv *Server) StartHttpServer() {
	if serv.http == nil {
		return
	}

	svr := serv.http
	console.LogKF("Http", "Running on http://localhost%s", svr.Addr)
	console.Fatal(serv.http.ListenAndServe())
}

func (serv *Server) Start() {
	go serv.StartHttpServer()

	v1.Banner()

	<-make(chan struct{})
}
`

const modelDbApi = `package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cgalvisleon/elvis/cache"
	"github.com/cgalvisleon/elvis/event"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/jrpc"
	"github.com/cgalvisleon/elvis/utility"
	"github.com/dimiro1/banner"
	"github.com/go-chi/chi/v5"
	"github.com/mattn/go-colorable"
	pkg "$1/pkg/$2"	
)

func New() http.Handler {
	r := chi.NewRouter()

	_, err := cache.Load()
	if err != nil {
		panic(err)
	}

	_, err = event.Load()
	if err != nil {
		panic(err)
	}

	db, err := jdb.Load()
	if err != nil {
		panic(err)
	}

	_pkg := &pkg.Router{
		Repository: &pkg.Controller{
			Db: db,
		},
	}

	r.Mount(pkg.PackagePath, _pkg.Routes())

	pkg.StartRpcServer()

	return r
}

func Close() {
	jrpc.Close()
	cache.Close()
	event.Close()
}

func Banner() {
	time.Sleep(3 * time.Second)
	templ := utility.BannerTitle(pkg.PackageName, pkg.PackageVersion, 4)
	banner.InitString(colorable.NewColorableStdout(), true, true, templ)
	fmt.Println()
}
`

const modelApi = `package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cgalvisleon/elvis/cache"
	"github.com/cgalvisleon/elvis/event"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/utility"
	"github.com/dimiro1/banner"
	"github.com/go-chi/chi/v5"
	"github.com/mattn/go-colorable"
	pkg "$1/pkg/$2"	
)

func New() http.Handler {
	r := chi.NewRouter()

	_, err := cache.Load()
	if err != nil {
		panic(err)
	}

	_, err = event.Load()
	if err != nil {
		panic(err)
	}

	db, err := jdb.Load()
	if err != nil {
		panic(err)
	}

	_pkg := &pkg.Router{
		Repository: &pkg.Controller{
			Db: db,
		},
	}

	r.Mount(pkg.PackagePath, _pkg.Routes())

	return r
}

func Close() {
	cache.Close()
	event.Close()
}

func Banner() {
	time.Sleep(3 * time.Second)
	templ := utility.BannerTitle(pkg.PackageName, pkg.PackageVersion, 4)
	banner.InitString(colorable.NewColorableStdout(), true, true, templ)
	fmt.Println()
}
`

const modelEvent = `package $1

import (
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/event"
	"github.com/cgalvisleon/elvis/et"
)

func initEvents() {
	err := event.Stack("<channel>", eventAction)
	if err != nil {
		console.Error(err)
	}

}

func eventAction(m event.EvenMessage) {
	data, err := et.ToJson(m.Data)
	if err != nil {
		console.Error(err)
	}

	console.Log("eventAction", data)
}
`

const modelModel = `package $1

import (
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/jdb"
)

func initModels(db *jdb.DB) error {
	if err := Define$2(db); err != nil {
		return console.Panic(err)
	}

	return nil
}
`

const modelSchema = `package $1

import (
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/linq"	
)

var $2 *linq.Schema

func defineSchema(db *jdb.DB) error {
	if $2 == nil {
		$2 = linq.NewSchema(db, "$3")
	}

	return nil
}
`

const modelhRpc = `package $1

import (
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/envar"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jrpc"
	"github.com/cgalvisleon/elvis/utility"
)

type Services struct{}

func StartRpcServer() {
	jrpc.Load()
	services := new(Services)
	err := jrpc.Mount(services, PackageName)
	if err != nil {
		console.Fatal(err)
	}

	go jrpc.StartServer()
}

func (c *Services) Version(require et.Json, response *et.Item) error {
	company := envar.GetStr("", "COMPANY")
	web := envar.GetStr("", "WEB")
	version := utility.GitVersion(1)
	help := envar.GetStr("", "RPC_HELP")
	response.Ok = true
	response.Result = et.Json{
		"methos":  "RPC",
		"version": version,
		"service": PackageName,
		"host":    HostName,
		"company": company,
		"web":     web,
		"help":    help,
	}

	return console.Rpc(response)
}

func (c *Services) Get$2ById(require et.Json, response *et.Item) error {
	id := require.Str("id")

	result, err := Get$2ById(id)
	if err != nil {
		return err
	}

	*response = result

	return console.Rpc(response)
}	
`

const modelMsg = `package $1

const (
	// MSG
	MSG_ATRIB_REQUIRED      = "Atributo requerido (%s)"
	MSG_VALUE_REQUIRED      = "Atributo requerido (%s) value:%s"
)
`

const modelDbController = `package $1

import (
	"context"

	"github.com/cgalvisleon/elvis/envar"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/utility"
)

type Controller struct {
	Db *jdb.DB
}

func (c *Controller) Version(ctx context.Context) (et.Json, error) {
	company := envar.GetStr("", "COMPANY")
	web := envar.GetStr("", "WEB")
	version := utility.GitVersion(1)
  service := et.Json{
		"version": version,
		"service": PackageName,
		"host":    HostName,
		"company": company,
		"web":     web,
		"help":    "",
	}

	return service, nil
}

func (c *Controller) Init(ctx context.Context) {
	initModels(c.Db)
	initEvents()
}

type Repository interface {
	Version(ctx context.Context) (et.Json, error)
	Init(ctx context.Context)
}
`

const modelController = `package $1

import (
	"context"

	"github.com/cgalvisleon/elvis/envar"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/utility"
)

type Controller struct {
	Db *jdb.DB
}

func (c *Controller) Version(ctx context.Context) (et.Json, error) {
	company := envar.GetStr("", "COMPANY")
	web := envar.GetStr("", "WEB")
	version := utility.GitVersion(1)
  service := et.Json{
		"version": version,
		"service": PackageName,
		"host":    HostName,
		"company": company,
		"web":     web,
		"help":    "",
	}

	return service, nil
}

func (c *Controller) Init(ctx context.Context) {
	initEvents()
}

type Repository interface {
	Version(ctx context.Context) (et.Json, error)
	Init(ctx context.Context)
}
`

const modelDbRouter = `package $1

import (
	"context"
	"net/http"
	"os"

	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/envar"
	"github.com/cgalvisleon/elvis/response"
	er "github.com/cgalvisleon/elvis/router"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
	"github.com/go-chi/chi/v5"
)

var PackageName = "$1"
var PackageTitle = "$1"
var PackagePath = envar.GetStr("/api/$1", "PATH_URL")
var PackageVersion = utility.GitVersion(0)
var HostName, _ = os.Hostname()

type Router struct {
	Repository Repository
}

func (rt *Router) Routes() http.Handler {
	defaultHost := strs.Format("http://%s", HostName)
	var host = strs.Format("%s:%d", envar.GetStr(defaultHost, "HOST"), envar.GetInt(3300, "PORT"))

	r := chi.NewRouter()

	er.PublicRoute(r, er.Get, "/version", rt.version, PackageName, PackagePath, host)
	// $2
	er.ProtectRoute(r, er.Get, "/{id}", rt.get$2ById, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Post, "/", rt.upSert$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Put, "/state/{id}", rt.state$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Delete, "/{id}", rt.delete$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Get, "/all", rt.all$2, PackageName, PackagePath, host)

	ctx := context.Background()
	rt.Repository.Init(ctx)

	console.LogKF(PackageName, "Router version:%s", PackageVersion)
	return r
}

func (rt *Router) version(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	result, err := rt.Repository.Version(ctx)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, result)
}
`

const modelRouter = `package $1

import (
	"context"
	"net/http"
	"os"

	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/envar"
	"github.com/cgalvisleon/elvis/response"
	er "github.com/cgalvisleon/elvis/router"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
	"github.com/go-chi/chi/v5"
)

var PackageName = "$1"
var PackageTitle = "$1"
var PackagePath = envar.GetStr("/api/$1", "PATH_URL")
var PackageVersion = utility.GitVersion(0)
var HostName, _ = os.Hostname()

type Router struct {
	Repository Repository
}

func (rt *Router) Routes() http.Handler {
	var host = strs.Format("%s:%d", envar.GetStr("http://localhost", "HOST"), envar.GetInt(3300, "PORT"))

	r := chi.NewRouter()

	er.PublicRoute(r, er.Get, "/version", rt.version, PackageName, PackagePath, host)
	// $2
	er.ProtectRoute(r, er.Post, "/", rt.$2, PackageName, PackagePath, host)
	
	ctx := context.Background()
	rt.Repository.Init(ctx)

	console.LogKF(PackageName, "Router version:%s", PackageVersion)
	return r
}

func (rt *Router) version(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	result, err := rt.Repository.Version(ctx)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, result)
}
`

const restHttp = `@host=localhost:3300
@token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IlVTRVIuQURNSU4iLCJhcHAiOiJEZXZvcHMtSW50ZXJuZXQiLCJuYW1lIjoiQ2VzYXIgR2FsdmlzIExlw7NuIiwia2luZCI6ImF1dGgiLCJ1c2VybmFtZSI6Iis1NzMxNjA0Nzk3MjQiLCJkZXZpY2UiOiJkZXZlbG9wIiwiZHVyYXRpb24iOjI1OTIwMDB9.dexIOute7r9o_P8U3t6l9RihN8BOnLl4xpoh9QbQI4k

###
GET /auth HTTP/1.1
Host: {{host}}/version
Authorization: Bearer {{token}}

###
POST /api/test/test HTTP/1.1
Host: {{host}}
Content-Type: application/json
Authorization: Bearer {{token}}
Content-Length: 227

{
}
`

const modelDbHandler = `package $1

import (
	"net/http"

	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/linq"
	"github.com/cgalvisleon/elvis/msg"
	"github.com/cgalvisleon/elvis/response"
	"github.com/cgalvisleon/elvis/utility"
	"github.com/go-chi/chi/v5"
)

var $2 *linq.Model

func Define$2(db *jdb.DB) error {
	if err := defineSchema(db); err != nil {
		return console.Panic(err)
	}

	if $2 != nil {
		return nil
	}

	$2 = linq.NewModel($3, "$4", "Tabla", 1)
	$2.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	$2.DefineColum("date_update", "", "TIMESTAMP", "NOW()")
	$2.DefineColum("_state", "", "VARCHAR(80)", utility.ACTIVE)
	$2.DefineColum("_id", "", "VARCHAR(80)", "-1")
	$2.DefineColum("project_id", "", "VARCHAR(80)", "-1")
	$2.DefineColum("name", "", "VARCHAR(250)", "")
	$2.DefineColum("description", "", "TEXT", "")
	$2.DefineColum("_data", "", "JSONB", "{}")
	$2.DefineColum("index", "", "INTEGER", 0)
	$2.DefinePrimaryKey([]string{"_id"})
	$2.DefineIndex([]string{
		"date_make",
		"date_update",
		"_state",
		"project_id",
		"name",
		"index",
	})
	$2.DefineRequired([]string{
		"name:Atributo requerido (name)",
	})
	$2.IntegrityAtrib(true)
	$2.IndexSource(true)
	$2.Trigger(linq.BeforeInsert, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		return nil
	})
	$2.Trigger(linq.AfterInsert, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		return nil
	})
	$2.Trigger(linq.BeforeUpdate, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		return nil
	})
	$2.Trigger(linq.AfterUpdate, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		return nil
	})
	$2.Trigger(linq.BeforeDelete, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		return nil
	})
	$2.Trigger(linq.AfterDelete, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		return nil
	})
	$2.OnListener = func(data et.Json) {
		console.Debug(data.ToString())
	}
	
	if err := $2.Init(); err != nil {
		return console.Panic(err)
	}

	return nil
}

/**
*	Get$2ById
* @param id string
* @return et.Item, error
**/
func Get$2ById(id string) (et.Item, error) {
	result, err := $2.Data().
		Where($2.Column("_id").Eq(id)).
		First()
	if err != nil {
		return et.Item{}, err
	}

	return result, nil	
}

/**
* Valida$2
* @params id, name string
* @return et.Item, error
**/
func Valida$2(id, name string) (et.Item, error) {
	item, err := $2.Data("_state", "_id").
		Where($2.Column("_id").Neg(id)).
		And($2.Column("name").Eq(name)).
		First()
	if err != nil {
		return et.Item{}, err
	}

	if item.Ok {
		return item, console.NewErrorF(msg.RECORD_DUPLICATE)
	}

	return et.Item{
		Ok: true,
	}, nil
}

/**
* Insert$2
* @params project_id, id, name, description string
* @params data et.Json
* @return et.Item, error
**/
func Insert$2(project_id, id, name, description string, data et.Json) (et.Item, error) {
	if !utility.ValidId(project_id) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "project_id")
	}

	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "service_id")
	}

	if !utility.ValidId(id) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "_id")
	}

	item, err := $2.Data("_state", "_id").
		Where($2.Column("_id").Eq(id)).
		First()
	if err != nil {
		return et.Item{}, err
	}

	if item.Ok {
		return et.Item{
			Ok:     false,
			Result: item.Result,
		}, nil
	}

	_, err = Valida$2(id, name)
	if err != nil {
		return et.Item{}, err
	}

	id = utility.GenId(id)
	now := utility.Now()
	data["date_make"] = now
	data["date_update"] = now
	data["project_id"] = project_id
	data["_id"] = id
	data["name"] = name
	data["description"] = description
	item, err = $2.Insert(data).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

/**
* UpSert$2
* @param project_id string
* @param id string
* @param data et.Json
* @return et.Item, error
**/
func UpSert$2(project_id, id, name, description string, data et.Json) (et.Item, error) {
	current, err := Insert$2(project_id, id, name, description, data)
	if err != nil {
		return et.Item{}, err
	}

	id = current.Key("_id")
	if !current.Ok {
		return Get$2ById(id)
	}

	current_state := current.Key("_state")
	if current_state != utility.ACTIVE {
		return et.Item{}, console.Alert(msg.RECORD_NOT_UPDATE)
	}

	_, err = Valida$2(id, name)
	if err != nil {
		return et.Item{}, err
	}

	id = utility.GenId(id)
	now := utility.Now()
	data["date_update"] = now
	data["project_id"] = project_id
	data["_id"] = id
	data["name"] = name
	data["description"] = description
	_, err = $2.Update(data).
		Where($2.Column("_id").Eq(id)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return Get$2ById(id)
}

/**
* State$2
* @param id, state string
* @return et.Item, error
**/
func State$2(id, state string) (et.Item, error) {
	if !utility.ValidId(state) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "state")
	}

	item, err := $2.Data("_state").
		Where($2.Column("_id").Eq(id)).
		First()
	if err != nil {
		return et.Item{}, err
	}

	if !item.Ok {
		return et.Item{}, console.Alert(msg.RECORD_NOT_FOUND)
	}

	old_state := item.Key("_state")
	if old_state == state {
		return et.Item{
			Ok: true,
			Result: et.Json{
				"message": msg.RECORD_NOT_UPDATE,
			}}, nil
	}

	return $2.Update(et.Json{
		"_state":   state,
	}).
		Where($2.Column("_id").Eq(id)).
		CommandOne()	
}

/**
* Delete$2
* @param id string
* @return et.Item, error
**/
func Delete$2(id string) (et.Item, error) {
	return State$2(id, utility.FOR_DELETE)
}

/**
* All$2
* @param project_id, state, search string
* @param page, rows int
* @param _select string
* @return et.List, error
**/
func All$2(project_id, state, search string, page, rows int, _select string) (et.List, error) {	
	if state == "" {
		state = utility.ACTIVE
	}

	auxState := state

	if search != "" {
		return $2.Data(_select).
			Where($2.Column("project_id").In("-1", project_id)).
			And($2.Concat("NAME:", $2.Column("name"), "DESCRIPTION:", $2.Column("description"), "DATA:", $2.Column("_data"), ":").Like("%"+search+"%")).
			OrderBy($2.Column("name"), true).
			List(page, rows)
	} else if auxState == "*" {
		state = utility.FOR_DELETE

		return $2.Data(_select).
			Where($2.Column("_state").Neg(state)).
			And($2.Column("project_id").In("-1", project_id)).
			OrderBy($2.Column("name"), true).
			List(page, rows)
	} else if auxState == "0" {
		return $2.Data(_select).
			Where($2.Column("_state").In("-1", state)).
			And($2.Column("project_id").In("-1", project_id)).
			OrderBy($2.Column("name"), true).
			List(page, rows)
	} else {
		return $2.Data(_select).
			Where($2.Column("_state").Eq(state)).
			And($2.Column("project_id").In("-1", project_id)).
			OrderBy($2.Column("name"), true).
			List(page, rows)
	}
}

/**
* upSert$2
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) upSert$2(w http.ResponseWriter, r *http.Request) {
	body, _ := response.GetBody(r)
	project_id := body.Str("project_id")
	id := body.Str("id")
	name := body.Str("name")
	description := body.Str("description")

	result, err := UpSert$2(project_id, id, name, description, body)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/**
* get$2ById
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) get$2ById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := Get$2ById(id)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/**
* state$2
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) state$2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	body, _ := response.GetBody(r)
	state := body.Str("state")

	result, err := State$2(id, state)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/**
* delete$2
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) delete$2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := Delete$2(id)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/**
* all$2
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) all$2(w http.ResponseWriter, r *http.Request) {
	query := response.GetQuery(r)
	project_id := query.Str("project_id")
	state := query.Str("state")
	search := query.Str("search")
	page := query.ValInt(1, "page")
	rows := query.ValInt(30, "rows")
	_select := query.Str("select")

	result, err := All$2(project_id, state, search, page, rows, _select)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, result)
}

/** Copy this code to router.go
	// $2
	er.ProtectRoute(r, er.Get, "/$5/{id}", rt.get$2ById, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Post, "/$5", rt.upSert$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Put, "/$5/state/{id}", rt.state$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Delete, "/$5/{id}", rt.delete$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Get, "/$5/all", rt.all$2, PackageName, PackagePath, host)
**/

/** Copy this code to func initModel in model.go
	if err := Define$2(db); err != nil {
		return console.Panic(err)
	}
**/
`

const modelHandler = `package $1

import (
	"net/http"

	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/response"
)

func $2(project_id, id string, params et.Json) (et.Item, error) {

	return et.Item{}, nil
}


/**
* Router
**/
func (rt *Router) $3(w http.ResponseWriter, r *http.Request) {
	body, _ := response.GetBody(r)
	project_id := body.Str("project_id")
	id := body.Str("id")	

	result, err := $2(project_id, id, body)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/** Copy this code to router.go
	// $2
	er.ProtectRoute(r, er.Post, "/$3", rt.$2, PackageName, PackagePath, Host)	
**/
`

const modelReadme = `
## Project $1

## Create project

go mod init github.com/$1/api

### Dependencias

go get github.com/cgalvisleon/elvis@v1.0.54
go get github.com/joho/godotenv/autoload
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/mongo/options
go get github.com/google/uuid
go get golang.org/x/crypto/bcrypt
go get golang.org/x/exp/slices
go get github.com/manifoldco/promptui
go get github.com/schollz/progressbar/v3
go get github.com/spf13/cobra
go get github.com/go-chi/chi/v5
go get github.com/redis/go-redis/v9
go get github.com/golang-jwt/jwt/v4
go get github.com/nats-io/nats.go
go get github.com/shirou/gopsutil/cpu
go get github.com/shirou/gopsutil/mem
go get github.com/lib/pq
go get github.com/dimiro1/banner
go get github.com/mattn/go-colorable
go get github.com/rs/cors

### Crear projecto, microservicios, modelos

go run github.com/cgalvisleon/elvis/cmd/create-go create

### Run project

gofmt -w . && go run ./cmd/$1 -port 3400 -rpc 4400
`

const modelEnvar = `APP=
PORT=3300
VERSION=0.0.0
COMPANY=Company
PATH_URL=
WEB=https://www.celsia.com
PRODUCTION=false
HOST=localhost

# DB
DB_DRIVE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=test
DB_USER=test
DB_PASSWORD=test
DB_APPLICATION_NAME=test

# REDIS
REDIS_HOST=localhost:6379
REDIS_PASSWORD=test
REDIS_DB=0

# NATS
NATS_HOST=localhost:4222

# CALM
SECRET=test

`

const modelDeploy = `version: "3"

networks:
  $3:
    external: true

services:
  $1:
    image: $1:latest
    logging:
      driver: "json-file"
      options:
        max-size: "1m"
        max-file: "2"
    networks:
      - $3
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.$1.rule=PathPrefix($2)"
      - "traefik.http.services.$1.loadbalancer.server.port=3300"
    deploy:
      replicas: 1
    environment:
      - "APP=Celsia Internet - Event Stack"
      - "PORT=3300"
      - "VERSION=1.0.1"
      - "COMPANY=Celsia Internet"
      - "WEB=https://www.internet.celsia.com"
      - "PATH_URL=/api/$1"
      - "PRODUCTION=true"
      - "HOST=stack"
      # DB
      - "DB_DRIVE=postgres"
      - "DB_HOST="
      - "DB_PORT=5432"
      - "DB_NAME=internet"
      - "DB_USER=internet"
      - "DB_PASSWORD="
      - "DB_APPLICATION_NAME=$1"
      # REDIS
      - "REDIS_HOST="
      - "REDIS_PASSWORD="
      - "REDIS_DB=0"
      # NATS
      - "NATS_HOST=nats:4222"
      # CALM
      - "SECRET="
      # RPC
      - "PORT_RPC=4200"
`

const modelGitignore = `# Created by https://www.toptal.com/developers/gitignore/api/go
# Edit at https://www.toptal.com/developers/gitignore?templates=go

### Go ###
# If you prefer the allow list template instead of the deny list, see community template:
# https://github.com/github/gitignore/blob/main/community/Golang/Go.AllowList.gitignore
#
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
.env
data
build
sql
.vscode

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

# Credencial acces token to AWS server
*.pem`
