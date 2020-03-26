package main

const qRelevance = `
CREATE TABLE %v (
    movie_id VARCHAR(10) PRIMARY KEY,
    expdate TIMESTAMP NOT NULL
)
`

const qUsers = `
CREATE TABLE %v (
    id SERIAL PRIMARY KEY,
    login CHAR(20) UNIQUE NOT NULL,
    password CHAR(128) NOT NULL
)
`

const qMovies = `
CREATE TABLE %v (
    id VARCHAR(10) PRIMARY KEY,
    title TEXT NOT NULL,
    year CHAR(9) NOT NULL,
    info TEXT NOT NULL
)
`
const qMovieToUser = `
CREATE TABLE %v (
    movie_id VARCHAR(10) NOT NULL,
    user_id INTEGER NOT NULL,
    PRIMARY KEY(movie_id, user_id)
)
`
