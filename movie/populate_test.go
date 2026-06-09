package movie

import "testing"

func TestPopulateMovieByLanguage_InsertPath(t *testing.T) {
	svc := NewService(nil, nil, nil)

	var inserted Movie
	insertCalled := 0
	updateCalled := 0
	personCalls := 0

	svc.getMovieByIdLanguageFn = func(id int, language string) Movie { return Movie{Id: 0} }
	svc.insertMovieFn = func(itemObj Movie, language string) interface{} {
		insertCalled++
		inserted = itemObj
		return nil
	}
	svc.updateMovieFn = func(itemObj Movie, language string) { updateCalled++ }
	svc.populatePersonFn = func(personId int, language, update string) { personCalls++ }

	movie := Movie{Id: 42, Title: "X"}
	movie.MovieCredits.Cast = []MovieCast{{Id: 1}, {Id: 2}}
	movie.MovieCredits.Crew = []MovieCrew{{Id: 3}}

	svc.PopulateMovieByLanguage(movie, "en", "N")

	if insertCalled != 1 || updateCalled != 0 {
		t.Fatalf("expected 1 insert, 0 update; got %d/%d", insertCalled, updateCalled)
	}
	if personCalls != 3 {
		t.Fatalf("expected 3 person calls, got %d", personCalls)
	}
	if inserted.Id != 42 || inserted.Language != "en" || inserted.MediaType != "movie" {
		t.Fatalf("unexpected inserted movie: %+v", inserted)
	}
}

func TestPopulateMovieByLanguage_UpdatePath(t *testing.T) {
	svc := NewService(nil, nil, nil)

	insertCalled := 0
	updateCalled := 0
	svc.getMovieByIdLanguageFn = func(id int, language string) Movie { return Movie{Id: 99} }
	svc.insertMovieFn = func(itemObj Movie, language string) interface{} { insertCalled++; return nil }
	svc.updateMovieFn = func(itemObj Movie, language string) { updateCalled++ }
	svc.populatePersonFn = func(personId int, language, update string) {
		t.Fatalf("populatePerson should not be called on update path")
	}

	svc.PopulateMovieByLanguage(Movie{Id: 99, Title: "Y"}, "pt-BR", "Y")

	if insertCalled != 0 || updateCalled != 1 {
		t.Fatalf("expected 0 insert, 1 update; got %d/%d", insertCalled, updateCalled)
	}
}

func TestPopulateMovieByLanguage_SkipPath(t *testing.T) {
	svc := NewService(nil, nil, nil)
	insertCalled := 0
	updateCalled := 0

	svc.getMovieByIdLanguageFn = func(id int, language string) Movie { return Movie{Id: 0} }
	svc.insertMovieFn = func(itemObj Movie, language string) interface{} { insertCalled++; return nil }
	svc.updateMovieFn = func(itemObj Movie, language string) { updateCalled++ }

	// itemObj.Id == 0 -> action = "skip"
	svc.PopulateMovieByLanguage(Movie{Id: 0}, "en", "N")

	if insertCalled != 0 || updateCalled != 0 {
		t.Fatalf("expected no insert/update; got %d/%d", insertCalled, updateCalled)
	}
}
