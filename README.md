# Moviedb Charge

[![Go Reference](https://pkg.go.dev/badge/golang.org/x/example.svg)](https://pkg.go.dev/golang.org/x/example)

This repository contains a Go program to get data from the [TMDB API](https://developer.themoviedb.org/docs "TMDB API") and insert/update in a MongoDb and an Elasticsearch instance to improve the search.

The main tasks (after the first charge) is to maintain a property database updated with the same data on the [TMDB API](https://developer.themoviedb.org/docs "TMDB API"). Using the daily export files to check if the catalog is complete and inserting new registers not found on the Charge database.

OBS.: The first charge needs to be used out of the schedule because it takes a long long time.
