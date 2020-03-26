# imdbCacher
Caches repsonses from imdb Alternative API (https://rapidapi.com/rapidapi/api/movie-database-imdb-alternative) to local db.

### How it works
You request a detailed information about some movie, server gives it to you by loading from imdb.
You request a detailed information about movie you requested before, server gives it to you by loading from local db.

### Requirements
- Configured PostgreSQL (1 user with password and db at least)

## Hot to run

### Server usage
```
go run imdbCacher/server -key <key> -user <name> -password <pass> -dbname <name> [-port <num>]
  -dbname string
    	db name
  -key string
    	(x-rapidapi-key) IMDB Alternative key
  -password string
    	db user password
  -port int
    	port to start server on (default 8080)
  -user string
    	db username

```
### Client usage
```
usage: go run imdbCacher/client -login|register -user <username> -password <pass> [-port <num>]
       go run imdbCacher/client -search <movie title> [-port <num>]
       go run imdbCacher/client -id <movie id> [-port <num>]
  -id string
    	search movie by imdb id returned after -search request
  -login
    	login operation
  -logout
    	logout operation
  -password string
    	user password
  -port int
    	port the server is listening on (default 8080)
  -register
    	user registeration
  -search string
    	search movie by title in imdb
  -user string
    	user name
```

### Workflow
1. Start server
2. Register/login using client (after registration you must login)
3. Use `-search` to get a movies list
4. Use `-id` with one of id's returned by `-search` to get a datailed movie info

## Examples of output
`-search`:
```
---------------
Title: Joker
Year: 2019
id: tt7286456
---------------
---------------
Title: Batman Beyond: Return of the Joker
Year: 2000
id: tt0233298
---------------
---------------
Title: Joker
Year: 2012
id: tt1918886
---------------
---------------
Title: Mera Naam Joker
Year: 1970
id: tt0066070
---------------

...
```
`-id`:
```
---------------
Title: The Joker Is Wild
Year: 1957
Genre: Biography, Drama, Musical
Actors: Frank Sinatra, Mitzi Gaynor, Jeanne Crain, Eddie Albert
Runtime: 126 min
Country: USA
Rating: 7.0
Type: movie
---------------

```

## Notes
`-search` requests aren't cached, only -id's are.
