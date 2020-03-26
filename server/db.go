package main

import (
    "os"
    "log"
    "fmt"
    "time"
    "errors"
    "strings"
    "strconv"
    "database/sql"
    "github.com/lib/pq"

    "imdbCacher/common"
)

// PUBLIC
type DBAccessor struct {
    db *sql.DB
}

func NewDBAcessor(driverName, dataSourceName string) (*DBAccessor, error) {
    var dba DBAccessor
    var err error
    dba.db, err = sql.Open(driverName, dataSourceName)

    if err != nil {
        return nil, err
    }

    err = dba.db.Ping()
    return &dba, err
}

func (dba *DBAccessor) Close() error {
    return dba.db.Close()
}

func (dba *DBAccessor) GetUser(login string) (*common.User, error) {
    const template = `SELECT login, password FROM %v WHERE login = %v`
    user := &common.User{}
    req, err := logAndGenQuery(template, tbUsers, login)

    if err != nil {
        return nil, err
    }

    err = dba.db.QueryRow(req, login).Scan(&user.Login, &user.Pass)
    user.Login = strings.Trim(user.Login, " ")
    user.Pass = strings.Trim(user.Pass, " ")
    return user, err
}

func (dba *DBAccessor) GetMovie(id string) (*Movie, error) {
    const template = `SELECT title, year, info FROM %v WHERE id = %v`
    var movie Movie
    req, err := logAndGenQuery(template, tbMovies, id)

    if err != nil {
        return nil, err
    }

    err = dba.db.QueryRow(req, id).Scan(&movie.Title, &movie.Year, &movie.Info)

    if err != nil {
        return nil, err
    }

    return &movie, err
}

func (dba *DBAccessor) CreateTablesIfNeeded() error {
    // tables are created atomically in scope
    // so we can check only existance of 1 table
    // and be sure others are present also
    //
    // except cases when user manualy modified db
    // which is not handled for now
    if res, err := dba.isTableExists(tbUsers); res || err != nil {
        return err
    }

    tx, err := dba.db.Begin()

    if err != nil {
        return err
    }

    defer tx.Rollback()
    tables := map[string]string {
        tbUsers: qUsers,
        tbMovies: qMovies,
        tbMovieToUser: qMovieToUser,
        tbRelevance: qRelevance,
    }

    for table, query := range tables {
        req := fmt.Sprintf(query, table)
        logDq.Printf(req)
        _, err = tx.Exec(req)

        if err != nil {
            return err
        }
    }

    err = tx.Commit()

    if err != nil {
        return err
    }

    return nil
}

func (dba *DBAccessor) AddUser(user common.User) error {
    const template = `INSERT INTO %v (login, password) VALUES(%v, %v)`
    req, err := logAndGenQuery(template, tbUsers, user.Login, user.Pass)

    if err != nil {
        return err
    }

    _, err = dba.db.Exec(req, user.Login, user.Pass)

    if err != nil {
        return err
    }

    return err
}

func (dba *DBAccessor) AddMovie(login string, movie Movie, expdate time.Time) error {
    userId, err := dba.getUserId(login)

    if err != nil {
        return err
    }

    tx, err := dba.db.Begin()

    if err != nil {
        return err
    }

    defer tx.Rollback()
    template := `INSERT INTO %v (id, title, year, info) VALUES(%v, %v, %v, %v)`
    req, err := logAndGenQuery(template, tbMovies,
            movie.ID, movie.Title, movie.Year, movie.Info)

    if err != nil {
        return err
    }

    _, err = tx.Exec(req, movie.ID, movie.Title, movie.Year, movie.Info)

    if err != nil {
        return err
    }

    template = `INSERT INTO %v (user_id, movie_id) VALUES(%v, %v)`
    req, err = logAndGenQuery(template, tbMovieToUser, strconv.Itoa(userId), movie.ID)

    if err != nil {
        return err
    }

    _, err = tx.Exec(req, userId, movie.ID)

    if err != nil {
        return err
    }

    template = `INSERT INTO %v (movie_id, expdate) VALUES(%v, %v)`
    timestamp := string(pq.FormatTimestamp(expdate))
    req, err = logAndGenQuery(template, tbRelevance, movie.ID, timestamp)

    if err != nil {
        return err
    }

    _, err = tx.Exec(req, movie.ID, timestamp)

    if err != nil {
        return err
    }

    err = tx.Commit()

    if err != nil {
        return err
    }

    return err
}

// PRIVATE
const (
    tbUsers string = "users"
    tbMovies string = "movies"
    tbRelevance string = "relevance"
    tbMovieToUser string = "movieToUser"
)

var logDq *log.Logger = log.New(os.Stdout, "[DEBUG QUERY]: ", log.Lshortfile)

func (dba *DBAccessor) isTableExists(tbname string) (bool, error) {
    const template = `SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = %v)`
    var exists bool
    logDq.Printf(template, `'` + tbname + `'`)
    req := fmt.Sprintf(template, "$1")
    err := dba.db.QueryRow(req, tbname).Scan(&exists)
    return exists, err
}

func (dba *DBAccessor) getUserId(name string) (int, error) {
    const template = `SELECT ID FROM %v WHERE login = %v`
    id := -1
    req, err := logAndGenQuery(template, tbUsers, name)

    if err != nil {
        return id, err
    }

    err = dba.db.QueryRow(req, name).Scan(&id)

    if err != nil {
        return id, err
    }

    return id, err
}

func logAndGenQuery(query string, args ...string) (string, error) {
    if len(args) < 2 {
        return "", errors.New("Insufficient number of arguments, must be 2 at least")
    }

    argsSize := len(args)
    prettifiedArgs := make([]interface{}, argsSize)
    resQueryArgs := make([]interface{}, argsSize)
    prettifiedArgs[0] = args[0]
    resQueryArgs[0] = args[0]

    for idx, arg := range args[1:] {
        idx += 1
        prettifiedArgs[idx] = `'` + arg + `'`
        resQueryArgs[idx] = "$" + strconv.Itoa(idx)
    }

    logDq.Printf(query, prettifiedArgs...)
    resQuery := fmt.Sprintf(query, resQueryArgs...)

    return resQuery, nil
}

