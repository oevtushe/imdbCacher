package main

import (
    "time"
    "fmt"
    "database/sql"
    "strings"
    "github.com/lib/pq"
)

const (
    tbUsers string = "users"
    tbMovies string = "movies"
    tbMovieToUser string = "movieToUser"
    tbRelevance string = "relevance"
)

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

func (dba *DBAccessor) GetUser(login string) (User, error) {
    req := fmt.Sprintf(`SELECT login, password FROM %v WHERE login = '%v'`,
                 tbUsers, login)
    dba.logQuery(req)
    var user User
    err := dba.db.QueryRow(req).Scan(&user.Login, &user.Pass)
    user.Login = strings.Trim(user.Login, " ")
    user.Pass = strings.Trim(user.Pass, " ")
    return user, err
}

func (dba *DBAccessor) GetMovie(id string) (*Movie, error) {
    req := fmt.Sprintf(`SELECT title, year, info FROM %v WHERE id = '%v'`,
                 tbMovies, id)
    dba.logQuery(req)
    var movie Movie
    err := dba.db.QueryRow(req).Scan(&movie.Title, &movie.Year, &movie.Info)

    if err != nil {
        return nil, err
    }

    return &movie, err
}

func (dba *DBAccessor) isTableExists(tbname string) (bool, error) {
    var exists bool
    req := fmt.Sprintf(`SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = '%v')`, tbname)
    dba.logQuery(req)
    err := dba.db.QueryRow(req).Scan(&exists)
    return exists, err
}

func (dba *DBAccessor) CreateTables() error {

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
        dba.logQuery(req)
        _, err = tx.Exec(req)

        if err != nil {
            return err
        }
    }

    err = tx.Commit()

    if err != nil {
        logE.Printf(err.Error())
        return err
    }

    return nil
}

func (dba DBAccessor) logQuery(query string) {
    logD.Printf("Executing query: %v\n", query)
}

func (dba *DBAccessor) AddUser(user User) error {
    req := fmt.Sprintf(`INSERT INTO %v (login, password) VALUES('%v', '%v')`,
                 tbUsers, user.Login, user.Pass)
    dba.logQuery(req)
    res, err := dba.db.Exec(req)

    if err != nil {
        return err
    }

    affected, err := res.RowsAffected()

    if err != nil {
        return err
    }

    logD.Printf("Rows affected (%v)\n", affected)

    return err
}

// TODO: may be needless
func (dba *DBAccessor) getUserId(name string) (int, error) {
    id := -1
    req := fmt.Sprintf(`SELECT ID FROM %v WHERE login = '%v'`, tbUsers, name)
    dba.logQuery(req)
    err := dba.db.QueryRow(req).Scan(&id)

    if err != nil {
        return id, err
    }

    return id, err
}

func (dba *DBAccessor) AddMovie(login string, movie Movie, expdate time.Time) error {
    userId, err := dba.getUserId(login)

    if err != nil {
        return err
    }

    var tx *sql.Tx
    tx, err = dba.db.Begin()

    if err != nil {
        return err
    }

    defer tx.Rollback()
    req := fmt.Sprintf(`INSERT INTO %v (id, title, year, info) VALUES('%v', '%v', '%v', '%v')`,
                    tbMovies, movie.ID, movie.Title, movie.Year, movie.Info)
    dba.logQuery(req)
    _, err = tx.Exec(req)

    if err != nil {
        return err
    }

    req = fmt.Sprintf(`INSERT INTO %v (user_id, movie_id) VALUES('%v', '%v')`,
            tbMovieToUser, userId, movie.ID)
    dba.logQuery(req)
    _, err = tx.Exec(req)

    if err != nil {
        return err
    }

    req = fmt.Sprintf(`INSERT INTO %v (movie_id, expdate) VALUES('%v', '%v')`,
            tbRelevance, movie.ID, string(pq.FormatTimestamp(expdate)))
    dba.logQuery(req)
    _, err = tx.Exec(req)

    if err != nil {
        return err
    }

    err = tx.Commit()

    if err != nil {
        return err
    }

    return err
}

/*
func main() {
    var dba *DBAccessor
    // TODO: do as options for program
    dba = NewDBAcessor("postgres", "user=test password=pass1 dbname=test")
    err := dba.db.Ping()

    if err != nil {
        logE.Fatal(err)
    }

    err = dba.CreateTables()

    if err != nil {
        logE.Fatal(err)
    }

    if err := dba.AddUser(User{"sasha", "pass"}); err != nil {
        logE.Fatal(err)
    }

    movie := imdbReq.Movie {
        Title: "Avengers Endgame",
        Year: "2019",
        ID: "iee228tf8",
    }

    if err := dba.AddMovie("sasha", movie, "additional info", time.Now()); err != nil {
        logE.Fatal(err)
    }

    user, err := dba.GetUser("oleg")

    if err == sql.ErrNoRows {
        logE.Printf("User not found !")
    }

    if err != nil {
        logE.Fatal(err)
    }

    fmt.Printf("%v\n", user.Login)

    defer dba.Close()
}
*/
